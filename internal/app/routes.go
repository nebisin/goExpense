package app

import (
	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/pkg/response"
	"net/http"
)

func (s *server) setupRoutes() {
	s.logger.Info("initializing the routes")

	s.router = mux.NewRouter()

	s.router.NotFoundHandler = http.HandlerFunc(response.NotFoundResponse)
	s.router.MethodNotAllowedHandler = http.HandlerFunc(response.MethodNotAllowedResponse)

	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/healthcheck", s.handleHealthCheck)

	apiV1.HandleFunc("/users", s.handleRegisterUser).Methods(http.MethodPost)
	apiV1.HandleFunc("/authenticate", s.handleLoginUser).Methods(http.MethodPost)
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