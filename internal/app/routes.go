package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/pkg/response"
)

func (s *server) setupRoutes() {
	s.logger.Info("initializing the routes")

	s.router = mux.NewRouter()

	s.router.Use(s.rateLimit)
	s.router.Use(s.authenticate)

	s.router.NotFoundHandler = http.HandlerFunc(response.NotFoundResponse)
	s.router.MethodNotAllowedHandler = http.HandlerFunc(response.MethodNotAllowedResponse)

	apiV1 := s.router.PathPrefix("/api/v1").Subrouter()

	apiV1.HandleFunc("/healthcheck", s.handleHealthCheck)

	apiV1.HandleFunc("/users", s.handleRegisterUser).Methods(http.MethodPost)
	apiV1.HandleFunc("/users", s.requireAuthenticatedUser(s.handleUpdateUser)).Methods(http.MethodPatch)
	apiV1.HandleFunc("/users/me", s.requireAuthenticatedUser(s.handleGetMe)).Methods(http.MethodGet)
	apiV1.HandleFunc("/users/accounts", s.requireAuthenticatedUser(s.handleGetAccounts)).Methods(http.MethodGet)
	apiV1.HandleFunc("/users/activate", s.handleActivateUser).Methods(http.MethodPut)
	apiV1.HandleFunc("/users/authenticate", s.handleLoginUser).Methods(http.MethodPost)
	apiV1.HandleFunc("/users/password", s.handlePasswordReset).Methods(http.MethodPut)

	apiV1.HandleFunc("/tokens/password-reset", s.handleCreatePasswordResetToken).Methods(http.MethodPost)
	apiV1.HandleFunc("/tokens/activation", s.handleNewActivationToken).Methods(http.MethodPost)

	apiV1.HandleFunc("/transactions", s.requireAuthenticatedUser(s.handleCreateTransaction)).Methods(http.MethodPost)
	apiV1.HandleFunc("/transactions/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleDeleteTransaction)).Methods(http.MethodDelete)
	apiV1.HandleFunc("/transactions/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleUpdateTransaction)).Methods(http.MethodPatch)
	apiV1.HandleFunc("/transactions/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleGetTransaction)).Methods(http.MethodGet)
	apiV1.HandleFunc("/transactions", s.requireAuthenticatedUser(s.handleListTransactions)).Methods(http.MethodGet)

	apiV1.HandleFunc("/accounts", s.requireAuthenticatedUser(s.handleCreateAccount)).Methods(http.MethodPost)
	apiV1.HandleFunc("/accounts/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleGetAccount)).Methods(http.MethodGet)
	apiV1.HandleFunc("/accounts/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleDeleteAccount)).Methods(http.MethodDelete)
	apiV1.HandleFunc("/accounts/{id:[0-9]+}", s.requireAuthenticatedUser(s.handleUpdateAccount)).Methods(http.MethodPatch)
	apiV1.HandleFunc("/accounts/{id:[0-9]+}/users", s.requireAuthenticatedUser(s.handleAddUser)).Methods(http.MethodPatch)
	apiV1.HandleFunc("/accounts/{id:[0-9]+}/users", s.requireAuthenticatedUser(s.handleGetUsers)).Methods(http.MethodGet)
	apiV1.HandleFunc("/accounts", s.requireAuthenticatedUser(s.handleListAccounts)).Methods(http.MethodGet)

	apiV1.HandleFunc("/accounts/{id:[0-9]+}/transactions", s.requireAuthenticatedUser(s.handleListTransactionsByAccount)).Methods(http.MethodGet)
	apiV1.HandleFunc("/accounts/{id:[0-9]+}/statistics", s.requireAuthenticatedUser(s.handleListStatisticsByAccount)).Methods(http.MethodGet)
}

func (s *server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if _, err := s.rdb.Ping(ctx).Result(); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	err := response.JSON(w, http.StatusOK, response.Envelope{
		"status":      "available",
		"environment": s.config.Env,
		"version":     version,
	})

	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
