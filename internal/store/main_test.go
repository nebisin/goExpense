package store_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/nebisin/goExpense/internal/store"
)

var testModels *store.Models
var testDB *sql.DB

func TestMain(m *testing.M) {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal(err)
	}

	db, err := store.OpenDB(os.Getenv("TEST_DB_URI"))
	if err != nil {
		log.Fatal(err)
	}

	testDB = db
	testModels = store.NewModels(db)

	os.Exit(m.Run())
}
