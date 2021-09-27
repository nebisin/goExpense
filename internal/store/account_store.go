package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Account struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	Version   int       `json:"version"`
}

type accountModel struct {
	DB *sql.DB
}

func (m *accountModel) Insert(account *Account) error {
	query := `INSERT INTO accounts (owner_id, title) 
VALUES ($1, $2) 
RETURNING id, created_at, version`

	args := []interface{}{
		account.OwnerID,
		account.Title,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&account.ID, &account.CreatedAt, &account.Version)
}

func (m *accountModel) Get(id int64) (*Account, error) {
	query := `SELECT id, owner_id, title, created_at, version
FROM accounts
WHERE id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var account Account

	err := m.DB.QueryRowContext(ctx, query, id).Scan(&account.ID, &account.OwnerID, &account.Title, &account.CreatedAt, &account.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRecordNotFound
	}

	return &account, err
}

func (m *accountModel) Delete(id int64, ownerID int64) error {
	query := `DELETE FROM accounts
WHERE id=$1 AND owner_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, ownerID)
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

func (m *accountModel) GetAll(ownerID int64, filters Filters) ([]*Account, error) {
	query := fmt.Sprintf(`SELECT id, owner_id, title, created_at, version
FROM accounts
WHERE owner_id=$1
ORDER BY %s %s, id ASC
LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, ownerID, filters.Limit, filters.offset())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []*Account{}

	for rows.Next() {
		var account Account

		err := rows.Scan(&account.ID, &account.OwnerID, &account.Title, &account.CreatedAt, &account.Version)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, &account)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return accounts, nil
}
