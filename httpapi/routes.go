package httpapi

import (
	"net/http"

	"github.com/gorilla/mux"
)

// API is the current API version
const API = "3.0"
const apiPath = "/api/" + API

// Router returns a new API router
func (s *Server) Router() http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix(apiPath).Subrouter()

	api.NotFoundHandler = withJSONResponse(func(r *http.Request) (int, interface{}) {
		return http.StatusNotFound, nil
	})

	api.Methods("POST").Path("/auth").Handler(
		withLogging("Authenticate", s.output,
			withJSONResponse(
				s.authenticate)))

	api.Methods("GET").Path("/users").Handler(
		withLogging("ListUsers", s.output,
			withJSONResponse(
				withAuth(s.sessionStore, s.listUsers))))

	api.Methods("POST").Path("/users/{username:[a-zA-Z]{2,6}[0-9]{1,2}}/reset").Handler(
		withLogging("ResetPassword", s.output,
			withJSONResponse(
				withAuth(s.sessionStore, s.resetPassword))))

	return r
}
