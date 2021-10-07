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

func (s *server) handleListStatisticsByAccount(w http.ResponseWriter, r *http.Request) {
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

	qs := r.URL.Query()

	before := request.ReadTime(qs, "before", time.Now())
	after := request.ReadTime(qs, "after", time.Now().AddDate(0, -1, 0))

	statistics, err := s.models.Statistics.GetAll(id, after, before)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"statistics": statistics}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
