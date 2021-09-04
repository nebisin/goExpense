package app

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/sirupsen/logrus"
	"os"
)

const version = "1.0.0"

type config struct {
	port  int
	env   string
	dbURI string
}

type server struct {
	router *mux.Router
	logger *logrus.Logger
	config config
	models *store.Models
}

func NewServer() *server {
	return &server{}
}

func (s *server) Run() {
	s.logger = logrus.New()
	s.logger.SetOutput(os.Stdout)
	s.logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if err := s.getConfig(); err != nil {
		s.logger.WithError(err).Fatal("something went wrong while getting env values")
	}

	s.logger.Info("we are connecting the database")
	db, err := store.OpenDB(s.config.dbURI)
	if err != nil {
		s.logger.WithError(err).Fatal("something went wrong while connecting the database")
	}
	defer db.Close()
	s.models = store.NewModels(db)

	s.setupRoutes()

	if err := s.serve(); err != nil {
		s.logger.WithError(err).Fatal("an error occurred while starting the server")
	}
}

func (s *server) getConfig() error {
	var cfg config

	s.logger.Info("we are getting env values")
	if err := godotenv.Load(); err != nil {
		return err
	}

	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.dbURI, "db-uri", os.Getenv("DB_URI"), "PostgreSQL DSN")

	flag.Parse()

	s.config = cfg

	return nil
}