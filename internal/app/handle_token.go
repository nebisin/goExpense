package app

import (
	"errors"
	"net/http"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
)

func (s *server) handleNewActivationToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email" validator:"required,email"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user, err := s.models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}

		return
	}

	if user.IsActivated {
		response.BadRequestResponse(w, r, errors.New("your account is already active"))
		return
	}

	if err := s.models.Tokens.DeleteAllForUser(store.ScopeActivation, user.ID); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	token, err := s.models.Tokens.New(user.ID, 3*24*time.Hour, store.ScopeActivation)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	s.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
		}

		if err := s.mailer.Send(user.Email, "new_token.tmpl", data); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"request_method": r.Method,
				"request_url":    r.URL.String(),
			}).WithError(err).Error("background email error")
		}
	})

	env := response.Envelope{"message": "an email will be sent to you containing the activation token"}

	if err := response.JSON(w, http.StatusAccepted, env); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleCreatePasswordResetToken(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email string `json:"email" validator:"required,email"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user, err := s.models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.NotFoundResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}

		return
	}

	if !user.IsActivated {
		response.FailedValidationResponse(w, r, map[string]string{"email": "user account must be activated"})
		return
	}

	token, err := s.models.Tokens.New(user.ID, 45*time.Minute, store.ScopePasswordReset)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	s.background(func() {
		data := map[string]interface{}{
			"passwordResetToken": token.Plaintext,
		}

		if err := s.mailer.Send(user.Email, "password_reset.tmpl", data); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"request_method": r.Method,
				"request_url":    r.URL.String(),
			}).WithError(err).Error("background email error")
		}
	})

	env := response.Envelope{"message": "an email will be sent to you containing password reset instructions"}

	if err := response.JSON(w, http.StatusAccepted, env); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
