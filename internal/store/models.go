package store

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrLongDuration   = errors.New("duration cannot be longer than one month")
)

type Models struct {
	Users        userModel
	Transactions transactionModel
	Tokens       tokenModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Users:        userModel{DB: db},
		Transactions: transactionModel{DB: db},
		Tokens:       tokenModel{DB: db},
	}
}
