package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Transaction struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userID"`
	AccountID   int64     `json:"accountID"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Amount      float64   `json:"amount"`
	Payday      time.Time `json:"payday"`
	CreatedAt   time.Time `json:"createdAt"`
	Version     int       `json:"version"`
	//Receipts    []string  `json:"receipts,omitempty"`
}

type transactionModel struct {
	DB *sql.DB
}

func (m *transactionModel) Insert(ts *Transaction) error {
	query := `INSERT INTO transactions (user_id, account_id, type, title, description, tags, amount, payday)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, version`

	args := []interface{}{
		ts.UserID,
		ts.AccountID,
		ts.Type,
		ts.Title,
		ts.Description,
		pq.Array(ts.Tags),
		ts.Amount,
		ts.Payday,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&ts.ID, &ts.CreatedAt, &ts.Version)
}

func (m *transactionModel) Get(id int64) (*Transaction, error) {
	query := `SELECT id, user_id, account_id, type, title, description, tags, amount, payday, created_at, version
FROM transactions
WHERE id = $1`

	var ts Transaction

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&ts.ID,
		&ts.UserID,
		&ts.AccountID,
		&ts.Type,
		&ts.Title,
		&ts.Description,
		pq.Array(&ts.Tags),
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
	query := `UPDATE transactions SET type=$1, title=$2, description=$3, tags=$4, amount=$5, payday=$6, version=version+1
WHERE id=$7 AND version=$8
RETURNING version`

	args := []interface{}{
		ts.Type,
		ts.Title,
		ts.Description,
		pq.Array(ts.Tags),
		ts.Amount,
		ts.Payday,
		ts.ID,
		ts.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&ts.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrEditConflict
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

func (m *transactionModel) GetAll(userId int64, title string, tags []string, startedAt time.Time, before time.Time, filters Filters) ([]*Transaction, error) {
	query := fmt.Sprintf(`SELECT id, user_id, account_id, type, title, description, tags, amount, payday, created_at, version
	FROM transactions
	WHERE user_id = $1 
	AND (to_tsvector('simple', title) @@ plainto_tsquery('simple', $2) OR $2='')
	OR (tags @> $7 OR $7 = '{}')
	AND payday >= $3 AND payday < $4
	ORDER BY %s %s, id ASC
	LIMIT $5 OFFSET $6`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		userId,
		title,
		startedAt,
		before,
		filters.Limit,
		filters.offset(),
		pq.Array(tags),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*Transaction{}

	for rows.Next() {
		var ts Transaction

		err := rows.Scan(
			&ts.ID,
			&ts.UserID,
			&ts.AccountID,
			&ts.Type,
			&ts.Title,
			&ts.Description,
			pq.Array(&ts.Tags),
			&ts.Amount,
			&ts.Payday,
			&ts.CreatedAt,
			&ts.Version,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, &ts)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

func (m *transactionModel) GetAllByAccountID(accountID int64, startedAt time.Time, before time.Time, filters Filters) ([]*Transaction, error) {
	query := fmt.Sprintf(`SELECT id, user_id, account_id, type, title, description, tags, amount, payday, created_at, version
FROM transactions
WHERE account_id=$1
AND payday >= $2 AND payday < $3
ORDER BY %s %s, id ASC 
LIMIT $4 OFFSET $5`, filters.sortColumn(), filters.sortDirection())

	args := []interface{}{
		accountID,
		startedAt,
		before,
		filters.Limit,
		filters.offset(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	transactions := []*Transaction{}

	for rows.Next() {
		var ts Transaction

		err := rows.Scan(
			&ts.ID,
			&ts.UserID,
			&ts.AccountID,
			&ts.Type,
			&ts.Title,
			&ts.Description,
			pq.Array(&ts.Tags),
			&ts.Amount,
			&ts.Payday,
			&ts.CreatedAt,
			&ts.Version,
		)
		if err != nil {
			return nil, err
		}

		transactions = append(transactions, &ts)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}
