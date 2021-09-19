package app

import (
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nebisin/goExpense/internal/mailer"
	"github.com/nebisin/goExpense/internal/store"
	"github.com/sirupsen/logrus"
)

const version = "1.0.0"

type config struct {
	port      int
	env       string
	dbURI     string
	jwtSecret string
	smtp      struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	cors struct {
		trustedOrigins []string
	}
}

type server struct {
	router *mux.Router
	logger *logrus.Logger
	config config
	db     *sql.DB
	models *store.Models
	wg     sync.WaitGroup
	mailer mailer.Mailer
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

	s.mailer = mailer.New(s.config.smtp.host, s.config.smtp.port, s.config.smtp.username, s.config.smtp.password, s.config.smtp.sender)

	s.logger.Info("we are connecting the database")
	db, err := store.OpenDB(s.config.dbURI)
	if err != nil {
		s.logger.WithError(err).Fatal("something went wrong while connecting the database")
	}
	defer db.Close()
	s.db = db
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
	flag.StringVar(&cfg.jwtSecret, "jwt-secret", os.Getenv("TOKEN_SYMMETRIC_KEY"), "JWT Secret")

	flag.StringVar(&cfg.smtp.host, "smtp-host", os.Getenv("SMTP_HOST"), "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", os.Getenv("SMTP_USERNAME"), "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", os.Getenv("SMTP_PASSWORD"), "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "goExpense <no-reply@goexpense.com>", "SMTP sender")

	flag.Func("cors-trusted-origins", "Trusted CORS origins (space seperated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	flag.Parse()

	s.config = cfg

	return nil
}
