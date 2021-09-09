package store

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type Models struct {
	Users        userModel
	Transactions transactionModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Users:        userModel{DB: db},
		Transactions: transactionModel{DB: db},
	}
}
