package store_test

import (
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomTransaction(t *testing.T) store.Transaction {
	user := createRandomUser(t)

	randTitle := random.String(12)
	randDesc := random.String(150)
	randTags := []string{random.String(6), random.String(6)}
	randAmount := float64(random.Int(1, 300))
	randPayday := time.Date(
		2020,
		time.December,
		int(random.Int(1, 29)),
		0,
		0,
		0,
		0,
		time.UTC,
	)

	var randType string
	if random.Int(1, 2) == 1 {
		randType = "income"
	} else {
		randType = "expense"
	}

	ts := store.Transaction{
		UserID:      user.ID,
		Type:        randType,
		Title:       randTitle,
		Description: randDesc,
		Tags:        randTags,
		Amount:      randAmount,
		Payday:      randPayday,
	}

	err := testModels.Transactions.Insert(&ts)
	require.NoError(t, err)
	require.NotEmpty(t, ts)

	require.Equal(t, ts.Title, randTitle)
	require.Equal(t, ts.Description, randDesc)
	require.Equal(t, ts.Tags, randTags)
	require.Equal(t, ts.Amount, randAmount)
	require.Equal(t, ts.UserID, user.ID)

	require.WithinDuration(t, ts.Payday, randPayday, time.Second)

	require.NotEmpty(t, ts.ID)
	require.NotEmpty(t, ts.Version)
	require.NotEmpty(t, ts.CreatedAt)

	return ts
}

func TestTransactionModel_Insert(t *testing.T) {
	createRandomTransaction(t)
}

func TestTransactionModel_Get(t *testing.T) {
	ts1 := createRandomTransaction(t)

	t.Run("success case for get transaction", func(t *testing.T) {
		ts2, err := testModels.Transactions.Get(ts1.ID)
		require.NoError(t, err)
		require.NotEmpty(t, ts2)

		require.Equal(t, ts1.Title, ts2.Title)
		require.Equal(t, ts1.Description, ts2.Description)
		require.Equal(t, ts1.Tags, ts2.Tags)
		require.Equal(t, ts1.Amount, ts2.Amount)
		require.Equal(t, ts1.UserID, ts2.UserID)

		require.WithinDuration(t, ts1.Payday, ts2.Payday, time.Second)
	})

	t.Run("not found case for get transaction", func(t *testing.T) {
		user2, err := testModels.Transactions.Get(ts1.ID + 1)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)
		require.Empty(t, user2)
	})

}
