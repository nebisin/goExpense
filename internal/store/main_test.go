package store_test

import (
	"github.com/nebisin/goExpense/pkg/config"
	"log"
	"os"
	"testing"

	"github.com/nebisin/goExpense/internal/store"
)

var testModels *store.Models

func TestMain(m *testing.M) {
	cfg, err := config.LoadConfig("../..", "test")
	if err != nil {
		log.Fatal(err)
	}

	db, err := store.OpenDB(cfg.DBURI)
	if err != nil {
		log.Fatal(err)
	}

	testModels = store.NewModels(db)

	os.Exit(m.Run())
}
