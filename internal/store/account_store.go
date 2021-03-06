package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Account struct {
	ID           int64     `json:"id"`
	OwnerID      int64     `json:"ownerID"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	TotalIncome  float64   `json:"totalIncome"`
	TotalExpense float64   `json:"totalExpense"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"createdAt"`
	Version      int       `json:"version"`
}

type accountModel struct {
	DB DBTX
}

func (m *accountModel) Insert(account *Account) error {
	query := `INSERT INTO accounts (owner_id, title, description, total_income, total_expense, currency) 
VALUES ($1, $2, $3, $4, $5, $6) 
RETURNING id, created_at, version`

	args := []interface{}{
		account.OwnerID,
		account.Title,
		account.Description,
		account.TotalIncome,
		account.TotalExpense,
		account.Currency,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&account.ID, &account.CreatedAt, &account.Version)
}

func (m *accountModel) Get(id int64) (*Account, error) {
	query := `SELECT id, owner_id, title, description, total_income, total_expense, currency, created_at, version
FROM accounts
WHERE id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var account Account

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.OwnerID,
		&account.Title,
		&account.Description,
		&account.TotalIncome,
		&account.TotalExpense,
		&account.Currency,
		&account.CreatedAt,
		&account.Version,
	)
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

func (m *accountModel) Update(account *Account) error {
	query := `UPDATE accounts SET title=$1, description=$2, total_income=$3, total_expense=$4, currency=$5, version=version+1
WHERE id=$6 AND version=$7
RETURNING version`

	args := []interface{}{
		account.Title,
		account.Description,
		account.TotalIncome,
		account.TotalExpense,
		account.Currency,
		account.ID,
		account.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&account.Version)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrEditConflict
	}

	return err
}

func (m *accountModel) GetAll(ownerID int64, filters Filters) ([]*Account, error) {
	query := fmt.Sprintf(`SELECT id, owner_id, title, description, total_income, total_expense, currency, created_at, version
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

		err := rows.Scan(
			&account.ID,
			&account.OwnerID,
			&account.Title,
			&account.Description,
			&account.TotalIncome,
			&account.TotalExpense,
			&account.Currency,
			&account.CreatedAt,
			&account.Version,
		)
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

func (m *accountModel) AddUser(userID int64, accountID int64) error {
	query := `INSERT INTO users_accounts (user_id, account_id)
	VALUES ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userID, accountID)

	return err
}

func (m *accountModel) RemoveUser(userID int64, accountID int64) error {
	query := `DELETE FROM users_accounts
	WHERE user_id=$1 AND account_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, userID, accountID)
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

func (m *accountModel) GetUsers(accountID int64) ([]*User, error) {
	query := `SELECT u.id, u.email, u.name, u.created_at, u.version
	FROM users_accounts a
	LEFT JOIN users u ON a.user_id = u.id
	WHERE account_id=$1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User

		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.CreatedAt,
			&user.Version,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
