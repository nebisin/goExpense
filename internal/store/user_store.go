package store

import (
	"database/sql"
	"time"
)

type User struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	HashedPassword []byte    `json:"-"`
	CreatedAt      time.Time `json:"created_at"`
	IsActivated    bool      `json:"is_activated"`
	Version        int       `json:"version"`
}

type userModel struct {
	DB *sql.DB
}
