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
		Type        string    `json:"type"`
		Title       string    `json:"title"`
		Description string    `json:"description,omitempty"`
		Amount      float64   `json:"amount"`
		Payday      time.Time `json:"payday"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
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

	if err := request.Validate(ts); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
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
		response.NotFoundResponse(w, r)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transaction": ts}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleDeleteTransaction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.NotFoundResponse(w, r)
		return
	}

	user := s.contextGetUser(r)

	if err := s.models.Transactions.Delete(id, user.ID); err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	err = response.JSON(w, http.StatusOK, response.Envelope{"message": "transaction successfully deleted"})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleUpdateTransaction(w http.ResponseWriter, r *http.Request) {
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
		response.NotFoundResponse(w, r)
		return
	}

	var input struct {
		Type        *string    `json:"type"`
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		Amount      *float64   `json:"amount"`
		Payday      *time.Time `json:"payday"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if input.Type != nil {
		ts.Type = *input.Type
	}

	if input.Title != nil {
		ts.Title = *input.Title
	}

	if input.Description != nil {
		ts.Description = *input.Description
	}

	if input.Amount != nil {
		ts.Amount = *input.Amount
	}

	if input.Payday != nil {
		ts.Payday = *input.Payday
	}

	if err := request.Validate(ts); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	if err := s.models.Transactions.Update(ts); err != nil {
		if errors.Is(err, store.ErrEditConflict) {
			response.EditConflictResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transaction": ts}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
