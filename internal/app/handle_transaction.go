package app

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
	"net/http"
	"strconv"
	"time"
)

func (s *server) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Type        string    `json:"type" validate:"required,oneof='expense' 'income'"`
		Title       string    `json:"title" validate:"required,max=180"`
		Description string    `json:"description,omitempty" validate:"max=1000"`
		Amount      float64   `json:"amount" validate:"required"`
		Payday      time.Time `json:"payday" validate:"required"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.ValidateInput(&input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user := s.contextGetUser(r)

	ts := &store.Transaction{
		UserID:      user.ID,
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		Amount:      input.Amount,
		Payday:      input.Payday,
	}

	if err := s.models.Transactions.Insert(ts); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	err := response.JSON(w, http.StatusCreated, response.Envelope{"transaction": ts})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}
}

func (s *server) handleGetTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.NotFoundResponse(w, r)
		return
	}

	ts, err := s.models.Transactions.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	user := s.contextGetUser(r)

	if ts.UserID != user.ID {
		response.NotPermittedResponse(w, r)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transaction": ts}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
