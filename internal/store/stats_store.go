package store

import (
	"context"
	"database/sql"
	"time"
)

type Stats struct {
	ID        int64     `json:"id"`
	AccountID int64     `json:"accountID"`
	Date      time.Time `json:"date"`
	Earning   float64   `json:"earning"`
	Spending  float64   `json:"spending"`
	CreatedAt time.Time `json:"createdAt"`
	Version   int       `json:"version"`
}

type statsModel struct {
	DB *sql.DB
}

func (m *statsModel) Insert(stats *Stats) error {
	query := `INSERT INTO stats (account_id, date, earning, spending)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`

	args := []interface{}{
		stats.AccountID,
		stats.Date,
		stats.Earning,
		stats.Spending,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&stats.ID, &stats.CreatedAt, &stats.Version)
}

func (m *statsModel) Get(id int64) (*Stats, error) {
	query := `SELECT id, account_id, date, earning, spending, created_at, version
	FROM stats
	WHERE id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stats Stats

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&stats.ID,
		&stats.AccountID,
		&stats.Earning,
		&stats.Spending,
		&stats.CreatedAt,
		&stats.Version,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}

	return &stats, err
}

func (m *statsModel) GetByDate(accountID int64, date time.Time) (*Stats, error) {
	query := `SELECT id, account_id, date, earning, spending, created_at, version
	FROM stats
	WHERE account_id = $1 AND data = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var stats Stats

	err := m.DB.QueryRowContext(ctx, query, accountID, date).Scan(
		&stats.ID,
		&stats.AccountID,
		&stats.Earning,
		&stats.Spending,
		&stats.CreatedAt,
		&stats.Version,
	)
	if err == sql.ErrNoRows {
		return nil, ErrRecordNotFound
	}

	return &stats, err
}

func (m *statsModel) Update(stats *Stats) error {
	query := `UPDATE stats SET earning=$1, spending=$2, version=version+1
	WHERE id=$3 AND version=$4
	RETURNING version`

	args := []interface{}{
		stats.Earning,
		stats.Spending,
		stats.ID,
		stats.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&stats.Version)
	if err == sql.ErrNoRows {
		return ErrRecordNotFound
	}

	return err
}

func (m *statsModel) Delete(id int64, accountID int64) error {
	query := `DELETE FROM stats
	WHERE id=$1 AND account_id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, accountID)
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

func (m *statsModel) GetAll(accountID int64, after time.Time, before time.Time) ([]*Stats, error) {
	query := `SELECT id, account_id, date, earning, spending, created_at, version
	FROM stats
	WHERE account_id=$1 AND date >= $2 AND date < $3
	ORDER BY id ASC`

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, accountID, after, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	statsList := []*Stats{}

	for rows.Next() {
		var stats Stats

		err := rows.Scan(
			&stats.ID,
			&stats.AccountID,
			&stats.Date,
			&stats.Earning,
			&stats.Spending,
			&stats.CreatedAt,
			&stats.Version,
		)
		if err != nil {
			return nil, err
		}
		statsList = append(statsList, &stats)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return statsList, nil
}
