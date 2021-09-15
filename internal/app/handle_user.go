package app

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/auth"
	"github.com/nebisin/goExpense/pkg/request"
	"github.com/nebisin/goExpense/pkg/response"
)

func (s *server) handleRegisterUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name" validate:"required,max=500"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,max=72,min=8"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(&input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user := &store.User{
		Name:        input.Name,
		Email:       input.Email,
		IsActivated: false,
	}

	if err := user.Password.Set(input.Password); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := s.models.Users.Insert(user); err != nil {
		if errors.Is(err, store.ErrDuplicateEmail) {
			errs := map[string]string{"email": "is already exist"}
			response.FailedValidationResponse(w, r, errs)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}

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
			"userID":          user.ID,
		}

		if err := s.mailer.Send(user.Email, "user_welcome.tmpl", data); err != nil {
			s.logger.WithFields(map[string]interface{}{
				"request_method": r.Method,
				"request_url":    r.URL.String(),
			}).WithError(err).Error("background email error")
		}
	})

	err = response.JSON(w, http.StatusCreated, response.Envelope{"user": user})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}
}

func (s *server) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,max=72,min=8"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(&input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user, err := s.models.Users.GetByEmail(input.Email)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.InvalidCredentialsResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	match, err := user.Password.Matches(input.Password)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if !match {
		response.InvalidCredentialsResponse(w, r)
		return
	}

	maker, err := auth.NewJWTMaker(s.config.jwtSecret)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	token, err := maker.CreateToken(user.ID, time.Hour*24*7)
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	err = response.JSON(w, http.StatusOK, response.Envelope{"authentication_token": token})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}

func (s *server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=500"`
		Email       *string `json:"email,omitempty" validate:"omitempty,email"`
		Password    *string `json:"password,omitempty" validate:"omitempty,max=72,min=8"`
		OldPassword *string `json:"old_password,omitempty" validate:"required_with=Password"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := request.Validate(input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user := s.contextGetUser(r)

	if input.Name != nil {
		user.Name = *input.Name
	}

	isEmailChanged := false
	if input.Email != nil && user.Email != *input.Email {
		isEmailChanged = true
		user.Email = *input.Email
		user.IsActivated = false
	}

	if input.Password != nil {
		match, err := user.Password.Matches(*input.OldPassword)
		if err != nil {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}

		if !match {
			response.InvalidCredentialsResponse(w, r)
			return
		}

		if err := user.Password.Set(*input.Password); err != nil {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}
	}

	if err := s.models.Users.Update(user); err != nil {
		if errors.Is(err, store.ErrDuplicateEmail) {
			errs := map[string]string{"email": "is already exist"}
			response.FailedValidationResponse(w, r, errs)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}

		return
	}

	if isEmailChanged {
		token, err := s.models.Tokens.New(user.ID, 3*24*time.Hour, store.ScopeActivation)
		if err != nil {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}

		// TODO: Send activation token via email
		fmt.Println(token.Plaintext) // delete after implementing activation email
	}

	err := response.JSON(w, http.StatusOK, response.Envelope{"user": user})
	if err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}
}

func (s *server) handleActivateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlainText string `json:"token" validator:"required,max=26"`
	}

	if err := request.ReadJSON(w, r, &input); err != nil {
		response.BadRequestResponse(w, r, err)
		return
	}

	if err := request.Validate(input); err != nil {
		response.FailedValidationResponse(w, r, err)
		return
	}

	user, err := s.models.Users.GetForToken(store.ScopeActivation, input.TokenPlainText)
	if err != nil {
		if errors.Is(err, store.ErrRecordNotFound) {
			response.FailedValidationResponse(w, r, map[string]string{"token": "invalid or expired activation token"})
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	user.IsActivated = true

	if err := s.models.Users.Update(user); err != nil {
		if errors.Is(err, store.ErrEditConflict) {
			response.EditConflictResponse(w, r)
		} else {
			response.ServerErrorResponse(w, r, s.logger, err)
		}
		return
	}

	if err := s.models.Tokens.DeleteAllForUser(store.ScopeActivation, user.ID); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
		return
	}

	if err := response.JSON(w, http.StatusOK, response.Envelope{"user": user}); err != nil {
		response.ServerErrorResponse(w, r, s.logger, err)
	}
}
