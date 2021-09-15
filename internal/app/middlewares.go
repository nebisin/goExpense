package app

import (
	"errors"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/auth"
	"github.com/nebisin/goExpense/pkg/response"
	"net/http"
	"strings"
)

func (s *server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = s.contextSetUser(r, store.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			response.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		maker, err := auth.NewJWTMaker(s.config.jwtSecret)
		if err != nil {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}

		payload, err := maker.VerifyToken(token)
		if err != nil {
			response.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := s.models.Users.Get(payload.UserID)
		if err != nil {
			if errors.Is(err, store.ErrRecordNotFound) {
				response.InvalidAuthenticationTokenResponse(w, r)
			} else {
				response.ServerErrorResponse(w, r, s.logger, err)
			}
			return
		}

		r = s.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}

func (s *server) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.contextGetUser(r)

		if user.IsAnonymous() {
			response.AuthenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *server) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.contextGetUser(r)

		if !user.IsActivated {
			response.InactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return s.requireAuthenticatedUser(fn)
}
