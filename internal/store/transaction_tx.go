package store

import (
	"context"
	"errors"
	"time"
)

func (m *Models) CreateTransactionTX(ts *Transaction, account *Account, statistic *Statistic) error {
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

	if ts.Type == "income" {
		account.TotalIncome += ts.Amount
	} else {
		account.TotalExpense += ts.Amount
	}

	if err := txModels.Accounts.Update(account); err != nil {
		return err
	}

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

func (m *Models) UpdateTransactionTX(newTS *Transaction, oldTS Transaction, account *Account, statistic *Statistic) error {
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
		if oldTS.Type == "income" {
			statistic.Earning -= oldTS.Amount - newTS.Amount
			account.TotalIncome -= oldTS.Amount - newTS.Amount
		} else {
			statistic.Spending -= oldTS.Amount - newTS.Amount
			account.TotalExpense -= oldTS.Amount - newTS.Amount
		}
	}

	if newTS.Type != oldTS.Type {
		if newTS.Type == "income" {
			statistic.Earning += newTS.Amount
			statistic.Spending -= newTS.Amount
			account.TotalIncome += newTS.Amount
			account.TotalExpense -= newTS.Amount
		} else {
			statistic.Earning -= newTS.Amount
			statistic.Spending += newTS.Amount
			account.TotalIncome -= newTS.Amount
			account.TotalExpense += newTS.Amount
		}
	}

	if oldTS.Payday != newTS.Payday {
		newStat, err := txModels.Statistics.GetByDate(newTS.AccountID, newTS.Payday)
		if err != nil {
			if errors.Is(err, ErrRecordNotFound) {
				newStat := &Statistic{}

				newStat.AccountID = newTS.AccountID
				newStat.Date = newTS.Payday

				if newTS.Type == "income" {
					statistic.Earning -= newTS.Amount
					newStat.Earning += newTS.Amount
				} else {
					statistic.Spending -= newTS.Amount
					newStat.Spending += newTS.Amount
				}

				if err := txModels.Statistics.Insert(newStat); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if newTS.Type == "income" {
				statistic.Earning -= newTS.Amount
				newStat.Earning += newTS.Amount
			} else {
				statistic.Spending -= newTS.Amount
				newStat.Spending += newTS.Amount
			}

			if err := txModels.Statistics.Update(newStat); err != nil {
				return err
			}
		}
	}

	if err := txModels.Accounts.Update(account); err != nil {
		return err
	}

	if err := txModels.Statistics.Update(statistic); err != nil {
		return err
	}

	return tx.Commit()
}

func (m *Models) DeleteTransactionTX(ts *Transaction, account *Account, stat *Statistic) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	txModels := NewModelsWithTX(tx)

	if err := txModels.Transactions.Delete(ts.ID, ts.UserID); err != nil {
		return err
	}

	if ts.Type == "income" {
		stat.Earning -= ts.Amount
		account.TotalIncome -= ts.Amount
	} else {
		stat.Spending -= ts.Amount
		account.TotalExpense -= ts.Amount
	}

	if err := txModels.Statistics.Update(stat); err != nil {
		return err
	}

	if err := txModels.Accounts.Update(account); err != nil {
		return err
	}

	return tx.Commit()
}
