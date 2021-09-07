package app

import (
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
	"net/http"
	"time"
)

func (s *server) handleCreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TSType      string  `json:"ts_type" validate:"required,oneof='expense' 'income'"`
		Title       string  `json:"title" validate:"required,max=180"`
		Description string  `json:"description,omitempty" validate:"max=1000"`
		Amount      float64 `json:"amount" validate:"required"`
		Payday      int64   `json:"payday" validate:"required"`
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
		TSType:      input.TSType,
		Title:       input.Title,
		Description: input.Description,
		Amount:      input.Amount,
		Payday:      time.UnixMilli(input.Payday),
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
