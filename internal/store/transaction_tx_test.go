package store_test

import (
	"testing"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
)

func createRandomTX(t *testing.T) (*store.Transaction, *store.Account, *store.Statistic) {
	user := createRandomUser(t)
	account := createRandomAccount(t)

	randTitle := random.String(12)
	randDesc := random.String(150)
	randTags := []string{random.String(6), random.String(6)}
	randAmount := float64(random.Int(1, 300))
	randPayday := random.Date()

	var randType string
	if random.Int(0, 2) == 1 {
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
	err := testModels.CreateTransactionTX(&ts, &account, &stat)
	require.NoError(t, err)
	require.NotEmpty(t, stat)

	if ts.Type == "income" {
		require.Equal(t, ts.Amount, stat.Earning)
		require.Equal(t, ts.Amount, account.TotalIncome)
	} else {
		require.Equal(t, ts.Amount, stat.Spending)
		require.Equal(t, ts.Amount, account.TotalExpense)
	}

	require.Equal(t, ts.AccountID, stat.AccountID)
	require.WithinDuration(t, ts.Payday, stat.Date, time.Second)

	return &ts, &account, &stat
}

func TestModels_CreateTransactionTX(t *testing.T) {
	createRandomTX(t)
}

func TestModels_UpdateTransactionTX(t *testing.T) {

	t.Run("update transaction amount test", func(t *testing.T) {
		oldTS, account, stat := createRandomTX(t)

		newTS := *oldTS

		newTS.Amount = float64(random.Int(1, 300))

		var expectedEarning float64
		var expectedSpending float64

		if oldTS.Type == "income" {
			expectedEarning = stat.Earning - (oldTS.Amount - newTS.Amount)
			expectedSpending = stat.Spending
		} else {
			expectedSpending = stat.Spending - (oldTS.Amount - newTS.Amount)
			expectedEarning = stat.Earning
		}

		err := testModels.UpdateTransactionTX(&newTS, *oldTS, account, stat)

		require.NoError(t, err)
		require.NotEmpty(t, newTS)
		require.NotEmpty(t, stat)

		require.Equal(t, stat.Earning, expectedEarning)
		require.Equal(t, stat.Spending, expectedSpending)

		require.Equal(t, account.TotalIncome, expectedEarning)
		require.Equal(t, account.TotalExpense, expectedSpending)
	})

	t.Run("update transaction type", func(t *testing.T) {
		oldTS, account, stat := createRandomTX(t)

		newTS := *oldTS

		var expectedEarning float64
		var expectedSpending float64

		if oldTS.Type == "income" {
			newTS.Type = "expense"
			expectedEarning = stat.Earning - newTS.Amount
			expectedSpending = stat.Spending + newTS.Amount
		} else {
			newTS.Type = "income"
			expectedEarning = stat.Earning + newTS.Amount
			expectedSpending = stat.Spending - newTS.Amount
		}

		err := testModels.UpdateTransactionTX(&newTS, *oldTS, account, stat)

		require.NoError(t, err)
		require.NotEmpty(t, newTS)
		require.NotEmpty(t, stat)

		require.Equal(t, stat.Earning, expectedEarning)
		require.Equal(t, stat.Spending, expectedSpending)

		require.Equal(t, account.TotalIncome, expectedEarning)
		require.Equal(t, account.TotalExpense, expectedSpending)
	})

	t.Run("update both type and amount", func(t *testing.T) {
		oldTS, account, stat := createRandomTX(t)

		newTS := *oldTS

		var expectedEarning float64
		var expectedSpending float64

		if oldTS.Type == "income" {
			expectedEarning = stat.Earning + (oldTS.Amount - newTS.Amount)
			expectedSpending = stat.Spending

			newTS.Type = "expense"
			expectedEarning -= newTS.Amount
			expectedSpending += newTS.Amount
		} else {
			expectedSpending = stat.Spending + (oldTS.Amount - newTS.Amount)
			expectedEarning = stat.Earning

			newTS.Type = "income"
			expectedEarning += newTS.Amount
			expectedSpending -= newTS.Amount
		}

		err := testModels.UpdateTransactionTX(&newTS, *oldTS, account, stat)

		require.NoError(t, err)
		require.NotEmpty(t, newTS)
		require.NotEmpty(t, stat)

		require.Equal(t, stat.Earning, expectedEarning)
		require.Equal(t, stat.Spending, expectedSpending)

		require.Equal(t, account.TotalIncome, expectedEarning)
		require.Equal(t, account.TotalExpense, expectedSpending)
	})

	t.Run("update transaction date", func(t *testing.T) {
		oldTS, account, stat := createRandomTX(t)

		newTS := *oldTS

		newTS.Payday = random.Date()

		var expectedEarning float64
		var expectedSpending float64
		if newTS.Type == "income" {
			expectedEarning = stat.Earning - newTS.Amount
			expectedSpending = stat.Spending
		} else {
			expectedEarning = stat.Earning
			expectedSpending = stat.Spending - newTS.Amount
		}

		err := testModels.UpdateTransactionTX(&newTS, *oldTS, account, stat)

		require.NoError(t, err)
		require.NotEmpty(t, newTS)
		require.NotEmpty(t, stat)

		require.Equal(t, stat.Earning, expectedEarning)
		require.Equal(t, stat.Spending, expectedSpending)
	})

}

func TestModels_DeleteTransactionTX(t *testing.T) {
	ts, account, stat := createRandomTX(t)

	var expectedEarning float64
	var expectedSpending float64

	if ts.Type == "income" {
		expectedEarning = stat.Earning - ts.Amount
		expectedSpending = stat.Spending
	} else {
		expectedEarning = stat.Earning
		expectedSpending = stat.Spending - ts.Amount
	}

	err := testModels.DeleteTransactionTX(ts, account, stat)
	require.NoError(t, err)
	require.NotEmpty(t, stat)

	require.Equal(t, stat.Earning, expectedEarning)
	require.Equal(t, stat.Spending, expectedSpending)
	require.Equal(t, account.TotalExpense, float64(0))
	require.Equal(t, account.TotalIncome, float64(0))
}
