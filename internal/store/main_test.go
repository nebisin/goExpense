package store_test

import (
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/nebisin/goExpense/internal/store"
)

var testModels *store.Models

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	db, err := store.OpenDB(os.Getenv("TEST_DB_URI"))
	if err != nil {
		log.Fatal(err)
	}

	testModels = store.NewModels(db)

	os.Exit(m.Run())
}
