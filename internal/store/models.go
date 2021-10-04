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
	Tokens       tokenModel
	Accounts     accountModel
	Stats        statsModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Users:        userModel{DB: db},
		Transactions: transactionModel{DB: db},
		Tokens:       tokenModel{DB: db},
		Accounts:     accountModel{DB: db},
		Stats:        statsModel{DB: db},
	}
}
