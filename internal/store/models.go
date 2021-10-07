package store

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
	ErrDuplicateEmail = errors.New("duplicate email")
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type Models struct {
	DB           *sql.DB
	Users        userModel
	Transactions transactionModel
	Tokens       tokenModel
	Accounts     accountModel
	Statistics   statisticModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		DB:           db,
		Users:        userModel{DB: db},
		Transactions: transactionModel{DB: db},
		Tokens:       tokenModel{DB: db},
		Accounts:     accountModel{DB: db},
		Statistics:   statisticModel{DB: db},
	}
}

func NewModelsWithTX(tx *sql.Tx) *Models {
	return &Models{
		Users:        userModel{DB: tx},
		Transactions: transactionModel{DB: tx},
		Tokens:       tokenModel{DB: tx},
		Accounts:     accountModel{DB: tx},
		Statistics:   statisticModel{DB: tx},
	}
}
