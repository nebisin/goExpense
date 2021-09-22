package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
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

	shutdownError := make(chan error)

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		sign := <-quit

		s.logger.WithField("signal", sign.String()).Info("shutting down the server")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		shutdownError <- srv.Shutdown(ctx)
	}()

	s.logger.WithField("port", s.config.port).Info("starting the server")

	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	s.logger.WithField("port", s.config.port).Info("stopped server")

	return nil
}
