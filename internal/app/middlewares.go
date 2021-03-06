package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/nebisin/goExpense/internal/cache"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/auth"
	"github.com/nebisin/goExpense/pkg/response"
	"golang.org/x/time/rate"
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

		maker, err := auth.NewJWTMaker(s.config.JwtSecret)
		if err != nil {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}

		payload, err := maker.VerifyToken(token)
		if err != nil {
			response.InvalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := s.cache.User.Get(payload.UserID)
		if err != nil && err != cache.ErrRecordNotFound {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}
		if user != nil {
			r = s.contextSetUser(r, user)
			next.ServeHTTP(w, r)
			return
		}

		user, err = s.models.Users.Get(payload.UserID)
		if err != nil {
			if errors.Is(err, store.ErrRecordNotFound) {
				response.InvalidAuthenticationTokenResponse(w, r)
			} else {
				response.ServerErrorResponse(w, r, s.logger, err)
			}
			return
		}

		s.background(func() {
			if err := s.cache.User.Set(user); err != nil {
				s.logger.WithFields(map[string]interface{}{
					"request_method": r.Method,
					"request_url":    r.URL.String(),
				}).WithError(err).Error("background cache error")
			}
		})

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

func (s *server) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" && len(s.config.CORS.TrustedOrigins) != 0 {
			for i := range s.config.CORS.TrustedOrigins {
				if origin == s.config.CORS.TrustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *server) rateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			response.ServerErrorResponse(w, r, s.logger, err)
			return
		}

		s.limiter.mu.Lock()

		if _, found := s.limiter.clients[ip]; !found {
			s.limiter.clients[ip] = &client{limiter: rate.NewLimiter(4, 6)}
		}

		s.limiter.clients[ip].lastSeen = time.Now()

		if !s.limiter.clients[ip].limiter.Allow() {
			s.limiter.mu.Unlock()
			response.RateLimitExceededResponse(w, r)
			return
		}

		s.limiter.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (s *server) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				response.ServerErrorResponse(w, r, s.logger, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}
