package app

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
	"net/http"
	"strconv"
)

func (s *server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var input struct{
		Name string `json:"name" validate:"required,min=3,max=500"`
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

	account := store.Account{
		UserID:    user.ID,
		Name:      input.Name,
	}

	err := s.models.Accounts.Insert(&account)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	err = response.JSON(w, http.StatusCreated, response.Envelope{"account": account})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleGetAccount(w http.ResponseWriter, r *http.Request) {
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

	if account.UserID != user.ID {
		response.NotFoundResponse(w, r)
		return
	}

	err = response.JSON(w, http.StatusOK, response.Envelope{"account": account})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.NotFoundResponse(w, r)
		return
	}

	user := s.contextGetUser(r)

	err = s.models.Accounts.Delete(id, user.ID)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	err = response.JSON(w, http.StatusOK, response.Envelope{"message": "account successfully deleted"})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleListAccounts(w http.ResponseWriter, r *http.Request) {
	var input struct{
		Filters store.Filters
	}

	qs := r.URL.Query()

	input.Filters.Page = request.ReadInt(qs, "page", 1)
	input.Filters.Limit = request.ReadInt(qs, "limit", 20)

	input.Filters.Sort = request.ReadString(qs, "sort", "id")

	if errs := request.Validate(input); errs != nil {
		response.FailedValidationResponse(w, r, errs)
		return
	}

	user := s.contextGetUser(r)

	accounts, err := s.models.Accounts.GetAll(user.ID, input.Filters)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	err = response.JSON(w, http.StatusOK, response.Envelope{"accounts": accounts})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}