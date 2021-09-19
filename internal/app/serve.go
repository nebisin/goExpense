package app

import (
	"fmt"
	"net/http"
	"time"
)

func (s *server) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.port),
		Handler:      s.recoverPanic(s.enableCORS(s.router)),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Minute,
	}

	s.logger.WithField("port", s.config.port).Info("starting the server")

	return srv.ListenAndServe()
}
