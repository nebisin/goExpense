package store_test

import (
	"testing"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) store.Account {
	user := createRandomUser(t)
	randomName := random.Name()

	account := store.Account{
		OwnerID:  user.ID,
		Title:    randomName,
		Currency: "USD",
	}

	err := testModels.Accounts.Insert(&account)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, account.OwnerID, user.ID)
	require.Equal(t, account.Title, randomName)
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
		require.Equal(t, account1.OwnerID, account2.OwnerID)
		require.Equal(t, account1.Title, account2.Title)
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

		err := testModels.Accounts.Delete(account1.ID, account1.OwnerID)
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

		err := testModels.Accounts.Delete(account1.ID, account1.OwnerID-1)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)

		account2, err := testModels.Accounts.Get(account1.ID)
		require.NoError(t, err)
		require.NotEmpty(t, account2)
	})
}

func TestAccountModel_Update(t *testing.T) {

	t.Run("success case for update account method", func(t *testing.T) {
		account1 := createRandomAccount(t)
		newDesc := random.Name()
		account2 := account1
		account2.Description = newDesc

		err := testModels.Accounts.Update(&account2)
		require.NoError(t, err)
		require.NotEmpty(t, account2)

		require.Equal(t, account2.ID, account1.ID)
		require.Equal(t, account2.OwnerID, account1.OwnerID)
		require.Equal(t, account2.Title, account1.Title)
		require.Equal(t, account2.Description, newDesc)
		require.Equal(t, account2.Version, account1.Version+1)

		require.WithinDuration(t, account2.CreatedAt, account1.CreatedAt, time.Second)
	})

	t.Run("edit conflict case for update account", func(t *testing.T) {
		account1 := createRandomAccount(t)
		account1.Version = int(random.Int(9, 99))

		err := testModels.Accounts.Update(&account1)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrEditConflict)
	})

}

func TestAccountModel_GetAll(t *testing.T) {
	account := createRandomAccount(t)

	accounts, err := testModels.Accounts.GetAll(account.OwnerID, store.Filters{
		Page:  1,
		Limit: 20,
		Sort:  "-id",
	})
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	require.Equal(t, len(accounts), 1)
	require.Equal(t, accounts[0].ID, account.ID)
	require.Equal(t, accounts[0].OwnerID, account.OwnerID)
	require.Equal(t, accounts[0].Title, account.Title)
	require.Equal(t, accounts[0].Version, account.Version)

	require.WithinDuration(t, accounts[0].CreatedAt, account.CreatedAt, time.Second)
}

func TestAccountModel_RemoveUser(t *testing.T) {
	account := createRandomAccount(t)

	err := testModels.Accounts.AddUser(account.OwnerID, account.ID)
	require.NoError(t, err)

	users, err := testModels.Accounts.GetUsers(account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, users)

	require.Equal(t, users[0].ID, account.OwnerID)

	err = testModels.Accounts.RemoveUser(account.OwnerID, account.ID)
	require.NoError(t, err)
}
