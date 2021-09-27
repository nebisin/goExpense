package store_test

import (
	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) store.Account {
	user := createRandomUser(t)
	randomName := random.Name()

	account := store.Account{
		UserID:    user.ID,
		Name:      randomName,
	}

	err := testModels.Accounts.Insert(&account)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, account.UserID, user.ID)
	require.Equal(t, account.Name, randomName)
	require.Equal(t, account.Version, 1)
	require.NotEmpty(t, account.CreatedAt)

	return account
}

func TestAccountModel_Insert(t *testing.T) {
	createRandomAccount(t)
}

func TestAccountModel_Get(t *testing.T) {
	account1 := createRandomAccount(t)

	t.Run("success case for get account method", func(t *testing.T) {
		account2, err := testModels.Accounts.Get(account1.ID)
		require.NoError(t, err)
		require.NotEmpty(t, account2)

		require.Equal(t, account1.ID, account2.ID)
		require.Equal(t, account1.UserID, account2.UserID)
		require.Equal(t, account1.Name, account2.Name)
		require.Equal(t, account1.Version, account2.Version)

		require.WithinDuration(t, account1.CreatedAt, account2.CreatedAt, time.Second)
	})

	t.Run("not found case for get account method", func(t *testing.T) {
		account2, err := testModels.Accounts.Get(random.Int(999, 9999))
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)
		require.Empty(t, account2)
	})
}

func TestAccountModel_Delete(t *testing.T) {
	t.Run("success case for delete account method", func(t *testing.T) {
		account1 := createRandomAccount(t)

		err := testModels.Accounts.Delete(account1.ID, account1.UserID)
		require.NoError(t, err)

		account2, err := testModels.Accounts.Get(account1.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)
		require.Empty(t, account2)
	})

	t.Run("not found case for delete account method", func(t *testing.T) {
		err := testModels.Accounts.Delete(random.Int(99, 999), random.Int(99, 999))
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)
	})

	t.Run("not authorized case for delete account method", func(t *testing.T) {
		account1 := createRandomAccount(t)

		err := testModels.Accounts.Delete(account1.ID, account1.UserID-1)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)

		account2, err := testModels.Accounts.Get(account1.ID)
		require.NoError(t, err)
		require.NotEmpty(t, account2)
	})
}