package app

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
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

func (s *server) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title string
		store.Filters
	}

	qs := r.URL.Query()

	input.Title = request.ReadString(qs, "title", "")

	input.Filters.Page = request.ReadInt(qs, "page", 1)
	input.Filters.Limit = request.ReadInt(qs, "limit", 20)

	input.Filters.Sort = request.ReadString(qs, "sort", "id")

	if errs := request.Validate(input); errs != nil {
		response.FailedValidationResponse(w, r, errs)
		return
	}

	user := s.contextGetUser(r)

	transactions, err := s.models.Transactions.GetAll(user.ID, input.Title, input.Filters)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transactions": transactions}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleListTransactionsByDate(w http.ResponseWriter, r *http.Request) {
	var input struct {
		startedAt time.Time
		before    time.Time
	}

	qs := r.URL.Query()

	input.before = request.ReadTime(qs, "before", time.Now())
	input.startedAt = request.ReadTime(qs, "started_at", time.Now().AddDate(0, -1, 0))

	user := s.contextGetUser(r)

	transactions, err := s.models.Transactions.GetAllByPayday(user.ID, input.startedAt, input.before)
	if err != nil {
		if errors.Is(err, store.ErrLongDuration) {
			response.BadRequestResponse(w, r, err)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}

		return
	}

	err = response.JSON(w, http.StatusOK, response.Envelope{"transactions": transactions})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
