package store

import (
	"context"
	"database/sql"
	"time"
)

type Statistic struct {
	AccountID int64     `json:"accountID"`
	Date      time.Time `json:"date"`
	Earning   float64   `json:"earning"`
	Spending  float64   `json:"spending"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int       `json:"version"`
}

type statisticModel struct {
	DB DBTX
}

func (m *statisticModel) Insert(stat *Statistic) error {
	query := `INSERT INTO statistics (account_id, date, earning, spending)
	VALUES ($1, $2, $3, $4)
	RETURNING created_at, version`

	args := []interface{}{
		stat.AccountID,
		stat.Date,
		stat.Earning,
		stat.Spending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&stat.CreatedAt, &stat.Version)
}

func (m *statisticModel) GetByDate(accountID int64, date time.Time) (*Statistic, error) {
	query := `SELECT account_id, date, earning, spending, created_at, version
	FROM statistics
	WHERE account_id = $1 AND date = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stat Statistic

	err := m.DB.QueryRowContext(ctx, query, accountID, date).Scan(
		&stat.AccountID,
		&stat.Earning,
		&stat.Spending,
		&stat.CreatedAt,
		&stat.Version,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}

	return &stat, err
}

func (m *statisticModel) Update(stat *Statistic) error {
	query := `UPDATE statistics SET earning=$1, spending=$2, version=version+1
	WHERE account_id=$3 AND date=$4 AND version=$5
	RETURNING version`

	args := []interface{}{
		stat.Earning,
		stat.Spending,
		stat.AccountID,
		stat.Date,
		stat.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&stat.Version)
	if err == sql.ErrNoRows {
		return ErrEditConflict
	}

	return err
}

func (m *statisticModel) Delete(accountID int64, date time.Time) error {
	query := `DELETE FROM statistics
	WHERE account_id=$1 AND date=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, accountID, date)
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

func (m *statisticModel) GetAll(accountID int64, after time.Time, before time.Time) ([]*Statistic, error) {
	query := `SELECT account_id, date, earning, spending, created_at, version
	FROM statistics
	WHERE account_id=$1 AND date >= $2 AND date < $3
	ORDER BY date ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, accountID, after, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []*Statistic{}

	for rows.Next() {
		var stat Statistic

		err := rows.Scan(
			&stat.AccountID,
			&stat.Date,
			&stat.Earning,
			&stat.Spending,
			&stat.CreatedAt,
			&stat.Version,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &stat)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}
