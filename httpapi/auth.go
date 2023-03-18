package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/korylprince/userbrowser-server/v3/session"
)

func (s *Server) authenticate(r *http.Request) (int, interface{}) {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	type response struct {
		Username    string `json:"username"`
		DisplayName string `json:"display_name"`
		SessionID   string `json:"session_id"`
	}

	req := new(request)

	if err := jsonRequest(r, req); err != nil {
		return http.StatusBadRequest, err
	}

	user, err := s.auth.Authenticate(req.Username, req.Password)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Unable to authenticate: %v", err)
	}

	if user == nil {
		return http.StatusUnauthorized, errors.New("Invalid username or password")
	}

	id, err := s.sessionStore.Create((*session.Session)(user))
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Unable to create session: %v", err)
	}

	return http.StatusOK, &response{
		Username:    user.Username,
		DisplayName: user.DisplayName,
		SessionID:   id,
	}
}

func withAuth(store session.Store, next returnHandlerFunc) returnHandlerFunc {
	return func(r *http.Request) (int, interface{}) {
		header := strings.Split(r.Header.Get("Authorization"), " ")

		if len(header) != 2 || header[0] != "Bearer" || len(header[1]) != 36 {
			return http.StatusBadRequest, errors.New("Invalid Authorization header")
		}

		session, err := store.Check(header[1])
		if err != nil {
			return http.StatusInternalServerError, fmt.Errorf("Unexpected error when checking session id %s: %v", header[1], err)
		}

		if session == nil {
			return http.StatusUnauthorized, fmt.Errorf("Session doesn't exist for id %s", header[1])
		}

		(r.Context().Value(contextKeyLogData)).(*logData).User = session.Username

		ctx := context.WithValue(r.Context(), contextKeyUser, session)

		status, body := next(r.WithContext(ctx))
		return status, body
	}
}
