package store

import "database/sql"

type Models struct {
	Users  userModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Users:  userModel{DB: db},
	}
}
