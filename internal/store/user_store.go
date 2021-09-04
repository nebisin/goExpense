package store

import (
	"context"
	"database/sql"
	"errors"
	"github.com/nebisin/goExpense/pkg/auth"
	"time"
)

type User struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	Email       string        `json:"email"`
	Password    auth.Password `json:"-"`
	CreatedAt   time.Time     `json:"created_at"`
	IsActivated bool          `json:"is_activated"`
	Version     int           `json:"version"`
}

type userModel struct {
	DB *sql.DB
}

func (m *userModel) Insert(user *User) error {
	query := `INSERT INTO users (name, email, hashed_password, is_activated)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version`

	args := []interface{}{user.Name, user.Email, user.Password.Hashed, user.IsActivated}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m *userModel) GetByEmail(email string) (*User, error) {
	query := `SELECT id, created_at, name, email, hashed_password, is_activated, version
FROM users
WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hashed,
		&user.IsActivated,
		&user.Version,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (m *userModel) Update(user *User) error {
	query := `UPDATE users
SET name = $1, email=$2, hashed_password=$3, is_activated=$4, version=version+1
WHERE id=$5 AND version=$6
RETURNING version`

	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.Hashed,
		user.IsActivated,
		user.ID,
		user.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrRecordNotFound
		default:
			return err
		}
	}

	return nil
}
