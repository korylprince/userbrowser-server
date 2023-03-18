package httpapi

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/korylprince/userbrowser-server/v3/auth"
	"github.com/korylprince/userbrowser-server/v3/db"
	"github.com/korylprince/userbrowser-server/v3/session"
)

type errResponse struct {
	Err string `json:"error"`
}

func (e *errResponse) Error() string {
	return e.Err
}

func (s *Server) listUsers(r *http.Request) (int, interface{}) {
	user := (*auth.User)((r.Context().Value(contextKeyUser)).(*session.Session))

	users, err := s.db.List()
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Unable to get user list: %v", err)
	}

	var filteredUsers []*db.User

	for _, u := range users {
		if user.Authorized(u.Grade) {
			filteredUsers = append(filteredUsers, u)
		}
	}

	return http.StatusOK, filteredUsers
}

func (s *Server) resetPassword(r *http.Request) (int, interface{}) {
	type response struct {
		Password string `json:"password"`
	}

	user := (*auth.User)((r.Context().Value(contextKeyUser)).(*session.Session))
	username := mux.Vars(r)["username"]

	resetUser, err := s.db.Get(username)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Unable locate user %s: %v", username, err)
	}

	if resetUser == nil || !user.Authorized(resetUser.Grade) {
		return http.StatusForbidden, fmt.Errorf("User %s doesn't have permissions to modify %s", user.Username, username)
	}

	(r.Context().Value(contextKeyLogData)).(*logData).ActionID = username

	passwd, err := s.db.ResetPassword(username)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Unable to reset password for user %s: %v", username, err)
	}

	return http.StatusOK, &response{Password: passwd}
}
