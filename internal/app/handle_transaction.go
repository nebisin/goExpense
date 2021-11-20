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
		AccountID   int64     `json:"accountID" validate:"required"`
		Type        string    `json:"type" validate:"required,oneof='expense' 'income'"`
		Title       string    `json:"title" validate:"required,min=3,max=180"`
		Description string    `json:"description,omitempty" validate:"max=1000"`
		Tags        []string  `json:"tags,omitempty" validate:"unique"`
		Amount      float64   `json:"amount" validate:"required"`
		Payday      time.Time `json:"payday" validate:"required"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user := s.contextGetUser(r)

	account, err := s.models.Accounts.Get(input.AccountID)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	// TODO: Update this while implementing the shared accounts
	if user.ID != account.OwnerID {
		response.NotPermittedResponse(w, r)
		return
	}

	ts := &store.Transaction{
		UserID:      user.ID,
		AccountID:   input.AccountID,
		Type:        input.Type,
		Title:       input.Title,
		Description: input.Description,
		Tags:        input.Tags,
		Amount:      input.Amount,
		Payday:      input.Payday,
	}

	stat, err := s.models.Statistics.GetByDate(account.ID, ts.Payday)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			stat = &store.Statistic{}
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}
	}

	if err := s.models.CreateTransactionTX(ts, account, stat); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	ts.Account = *account
	ts.User = *user

	err = response.JSON(w, http.StatusCreated, response.Envelope{"transaction": ts})
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

	ts.User = *user

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

	ts, err := s.models.Transactions.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	if ts.UserID != user.ID {
		response.NotFoundResponse(w, r)
		return
	}

	stat, err := s.models.Statistics.GetByDate(ts.AccountID, ts.Payday)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	account, err := s.models.Accounts.Get(ts.AccountID)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := s.models.DeleteTransactionTX(ts, account, stat); err != nil {
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

	oldTS, err := s.models.Transactions.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	user := s.contextGetUser(r)

	if oldTS.UserID != user.ID {
		response.NotFoundResponse(w, r)
		return
	}

	var input struct {
		Type        *string    `json:"type,omitempty" validate:"omitempty,oneof='expense' 'income'"`
		Title       *string    `json:"title,omitempty" validate:"omitempty,min=3,max=180"`
		Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
		Tags        []string   `json:"tags,omitempty" validate:"unique"`
		Amount      *float64   `json:"amount,omitempty"`
		Payday      *time.Time `json:"payday,omitempty"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	newTS := *oldTS

	if input.Type != nil {
		newTS.Type = *input.Type
	}

	if input.Title != nil {
		newTS.Title = *input.Title
	}

	if input.Description != nil {
		newTS.Description = *input.Description
	}

	if input.Tags != nil {
		newTS.Tags = input.Tags
	}

	if input.Amount != nil {
		newTS.Amount = *input.Amount
	}

	if input.Payday != nil {
		newTS.Payday = *input.Payday
	}

	stat, err := s.models.Statistics.GetByDate(oldTS.AccountID, oldTS.Payday)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	account, err := s.models.Accounts.Get(oldTS.AccountID)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := s.models.UpdateTransactionTX(&newTS, *oldTS, account, stat); err != nil {
		if errors.Is(err, store.ErrEditConflict) {
			response.EditConflictResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transaction": newTS}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleListTransactions(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title     string
		Tags      []string
		Before    time.Time
		StartedAt time.Time
		store.Filters
	}

	qs := r.URL.Query()

	input.Title = request.ReadString(qs, "title", "")
	input.Tags = request.ReadCSV(qs, "tags", []string{})

	input.Filters.Page = request.ReadInt(qs, "page", 1)
	input.Filters.Limit = request.ReadInt(qs, "limit", 20)

	input.Filters.Sort = request.ReadString(qs, "sort", "id")

	input.Before = request.ReadTime(qs, "before", time.Now().AddDate(3, 0, 0))
	input.StartedAt = request.ReadTime(qs, "startedAt", time.Unix(0, 0))

	if errs := request.Validate(input); errs != nil {
		response.FailedValidationResponse(w, r, errs)
		return
	}

	user := s.contextGetUser(r)

	transactions, err := s.models.Transactions.GetAll(user.ID, input.Title, input.Tags, input.StartedAt, input.Before, input.Filters)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transactions": transactions}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleListTransactionsByAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.NotFoundResponse(w, r)
		return
	}

	account, err := s.models.Accounts.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	user := s.contextGetUser(r)

	// TODO: Update this while implementing the shared accounts
	if user.ID != account.OwnerID {
		response.NotPermittedResponse(w, r)
		return
	}

	var input struct {
		Before    time.Time
		StartedAt time.Time
		store.Filters
	}

	qs := r.URL.Query()

	input.Filters.Page = request.ReadInt(qs, "page", 1)
	input.Filters.Limit = request.ReadInt(qs, "limit", 20)

	input.Filters.Sort = request.ReadString(qs, "sort", "id")

	input.Before = request.ReadTime(qs, "before", time.Now().AddDate(3, 0, 0))
	input.StartedAt = request.ReadTime(qs, "startedAt", time.Unix(0, 0))

	if errs := request.Validate(input); errs != nil {
		response.FailedValidationResponse(w, r, errs)
		return
	}

	transactions, err := s.models.Transactions.GetAllByAccountID(account.ID, input.StartedAt, input.Before, input.Filters)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"transactions": transactions}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
