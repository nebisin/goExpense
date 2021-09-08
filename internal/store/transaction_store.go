package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Transaction struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"user_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Amount      float64   `json:"amount"`
	Payday      time.Time `json:"payday"`
	CreatedAt   time.Time `json:"created_at"`
	Version     int       `json:"version"`
	//AccountID   int64     `json:"account_id,omitempty"`
	//Tags        []string  `json:"tags,omitempty"`
	//Receipts    []string  `json:"receipts,omitempty"`
}

type transactionModel struct {
	DB *sql.DB
}

func (m *transactionModel) Insert(ts *Transaction) error {
	query := `INSERT INTO transactions (user_id, type, title, description, amount, payday)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at, version`

	args := []interface{}{
		ts.UserID,
		ts.Type,
		ts.Title,
		ts.Description,
		ts.Amount,
		ts.Payday,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&ts.ID, &ts.CreatedAt, &ts.Version)
}

func (m *transactionModel) Get(id int64) (*Transaction, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT id, user_id, type, title, description, amount, payday, created_at, version
FROM transactions
WHERE id = $1`

	var ts Transaction

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&ts.ID,
		&ts.UserID,
		&ts.Type,
		&ts.Title,
		&ts.Description,
		&ts.Amount,
		&ts.Payday,
		&ts.CreatedAt,
		&ts.Version,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecordNotFound
	}

	return &ts, err
}

// TODO: Create update, delete and getAll methods
