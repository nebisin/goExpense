package store

import (
	"context"
	"time"
)

func (m *Models) CreateAccountTX(account *Account) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txModels := NewModelsWithTX(tx)

	err = txModels.Accounts.Insert(account)
	if err != nil {
		return err
	}

	err = txModels.Accounts.AddUser(account.OwnerID, account.ID)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil

}
