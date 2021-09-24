package app

import (
	"database/sql"
	"github.com/nebisin/goExpense/pkg/config"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/nebisin/goExpense/internal/mailer"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

const version = "1.0.0"

type server struct {
	router  *mux.Router
	logger  *logrus.Logger
	config  config.Config
	db      *sql.DB
	models  *store.Models
	wg      sync.WaitGroup
	mailer  mailer.Mailer
	limiter struct {
		mu      sync.Mutex
		clients map[string]*client
	}
}

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewServer() *server {
	return &server{}
}

func (s *server) Run() {
	s.logger = logrus.New()
	s.logger.SetOutput(os.Stdout)
	s.logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	cfg, err := config.LoadConfig(".", "app")
	if err != nil {
		s.logger.WithError(err).Fatal("something went wrong while getting env values")
	}
	s.config = cfg

	s.mailer = mailer.New(s.config.SMTP.Host, s.config.SMTP.Port, s.config.SMTP.Username, s.config.SMTP.Password, s.config.SMTP.Sender)

	s.logger.Info("we are connecting the database")
	db, err := store.OpenDB(s.config.DbURI)
	if err != nil {
		s.logger.WithError(err).Fatal("something went wrong while connecting the database")
	}
	defer db.Close()
	s.db = db
	s.models = store.NewModels(db)

	s.setupRoutes()

	s.setupLimiter()

	if err := s.serve(); err != nil {
		s.logger.WithError(err).Fatal("an error occurred while starting the server")
	}
}

func (s *server) setupLimiter() {
	s.limiter.clients = make(map[string]*client)

	go func() {
		for {
			time.Sleep(time.Minute)

			s.limiter.mu.Lock()

			for ip, client := range s.limiter.clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(s.limiter.clients, ip)
				}
			}

			s.limiter.mu.Unlock()
		}
	}()
}
