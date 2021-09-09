package app

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/pkg/response"
)

func (s *server) setupRoutes() {
	s.logger.Info("initializing the routes")

	s.router = mux.NewRouter()

	s.router.Use(s.authenticate)

	s.router.NotFoundHandler = http.HandlerFunc(response.NotFoundResponse)
	s.router.MethodNotAllowedHandler = http.HandlerFunc(response.MethodNotAllowedResponse)

	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/healthcheck", s.handleHealthCheck)

	apiV1.HandleFunc("/users", s.handleRegisterUser).Methods(http.MethodPost)
	apiV1.HandleFunc("/authenticate", s.handleLoginUser).Methods(http.MethodPost)

	apiV1.HandleFunc("/transactions", s.requireAuthenticatedUser(s.handleCreateTransaction)).Methods(http.MethodPost)
	apiV1.HandleFunc("/transactions/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleGetTransaction)).Methods(http.MethodGet)
	apiV1.HandleFunc("/transactions/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleDeleteTransaction)).Methods(http.MethodDelete)
	apiV1.HandleFunc("/transactions/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleUpdateTransaction)).Methods(http.MethodPatch)
	apiV1.HandleFunc("/transactions", s.requireAuthenticatedUser(s.handleListTransactions)).Methods(http.MethodGet)
}

func (s *server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	err := response.JSON(w, http.StatusOK, response.Envelope{
		"status":      "available",
		"environment": s.config.env,
		"version":     version,
	})

	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
