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
	Type        string    `json:"type" validate:"required,oneof='expense' 'income'"`
	Title       string    `json:"title" validate:"required,max=180"`
	Description string    `json:"description,omitempty" validate:"max=1000"`
	Amount      float64   `json:"amount" validate:"required"`
	Payday      time.Time `json:"payday" validate:"required"`
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

func (m *transactionModel) Update(ts *Transaction) error {
	query := `UPDATE transactions SET type=$1, title=$2, description=$3, amount=$4, payday=$5, version=version+1
WHERE id=$6 AND version=$7
RETURNING version`

	args := []interface{}{
		ts.Type,
		ts.Title,
		ts.Description,
		ts.Amount,
		ts.Payday,
		ts.ID,
		ts.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&ts.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrRecordNotFound
	}

	return err
}

func (m *transactionModel) Delete(id int64, userID int64) error {
	query := `DELETE FROM transactions
WHERE id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// TODO: Implement GetAll method

func (m *transactionModel) GetAll(title string, filters Filters) ([]*Transaction, error) {
	return nil, nil
}
