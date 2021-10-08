package store

import (
	"context"
	"time"
)

func (m *Models) CreateTransactionTX(ts *Transaction, statistic *Statistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txModels := NewModelsWithTX(tx)

	if err := txModels.Transactions.Insert(ts); err != nil {
		return err
	}

	// TODO: Update account balance
	/*
		if err := txModels.Accounts.Update(account); err != nil {
			return err
		}
	*/
	if statistic.Version == 0 {
		statistic.AccountID = ts.AccountID
		statistic.Date = ts.Payday

		if ts.Type == "income" {
			statistic.Earning = ts.Amount
		} else {
			statistic.Spending = ts.Amount
		}

		if err := txModels.Statistics.Insert(statistic); err != nil {
			return err
		}
	} else {
		if ts.Type == "income" {
			statistic.Earning += ts.Amount
		} else {
			statistic.Spending += ts.Amount
		}

		if err := txModels.Statistics.Update(statistic); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *Models) UpdateTransactionTX(newTS *Transaction, oldTS Transaction, statistic *Statistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txModels := NewModelsWithTX(tx)

	if err := txModels.Transactions.Update(newTS); err != nil {
		return err
	}

	if newTS.Amount != oldTS.Amount {
		if newTS.Type == "income" {
			statistic.Earning += oldTS.Amount - newTS.Amount
		} else {
			statistic.Spending += oldTS.Amount - newTS.Amount
		}
	}

	if newTS.Type != oldTS.Type {
		if newTS.Type == "income" {
			statistic.Earning += newTS.Amount
			statistic.Spending -= newTS.Amount
		} else {
			statistic.Earning -= newTS.Amount
			statistic.Spending += newTS.Amount
		}
	}

	// TODO: Cover payday changing

	if err := txModels.Statistics.Update(statistic); err != nil {
		return err
	}

	return tx.Commit()
}
