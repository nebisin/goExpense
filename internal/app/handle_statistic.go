package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
)

func (s *server) handleListStatisticsByAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		response.NotFoundResponse(w, r)
		return
	}

	user := s.contextGetUser(r)
	users, err := s.models.Accounts.GetUsers(id)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	isMember := false
	for _, value := range users {
		if value.ID == user.ID {
			isMember = true
			break
		}
	}
	if !isMember {
		response.NotFoundResponse(w, r)
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
