package store_test

import (
	"testing"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
)

func TestModels_CreateTransactionTX(t *testing.T) {
	user := createRandomUser(t)
	account := createRandomAccount(t)

	randTitle := random.String(12)
	randDesc := random.String(150)
	randTags := []string{random.String(6), random.String(6)}
	randAmount := float64(random.Int(1, 300))
	randPayday := random.Date()

	var randType string
	if random.Int(1, 2) == 1 {
		randType = "income"
	} else {
		randType = "expense"
	}

	ts := store.Transaction{
		UserID:      user.ID,
		AccountID:   account.ID,
		Type:        randType,
		Title:       randTitle,
		Description: randDesc,
		Tags:        randTags,
		Amount:      randAmount,
		Payday:      randPayday,
	}

	stat := store.Statistic{}
	err := testModels.CreateTransactionTX(&ts, &stat)
	require.NoError(t, err)
	require.NotEmpty(t, stat)

	if ts.Type == "income" {
		require.Equal(t, ts.Amount, stat.Earning)
	} else {
		require.Equal(t, ts.Amount, stat.Spending)
	}

	require.Equal(t, ts.AccountID, stat.AccountID)
	require.WithinDuration(t, ts.Payday, stat.Date, time.Second)
}
