package store_test

import (
	"testing"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) store.User {
	randomName := random.Name()
	randomEmail := random.Email()
	randomPassword := random.Password()

	user := store.User{
		Name:        randomName,
		Email:       randomEmail,
		IsActivated: false,
	}

	err := user.Password.Set(randomPassword)
	require.NoError(t, err)

	err = testModels.Users.Insert(&user)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, randomName, user.Name)
	require.Equal(t, randomEmail, user.Email)
	require.Equal(t, false, user.IsActivated)
	require.Equal(t, 1, user.Version)

	ok, err := user.Password.Matches(randomPassword)
	require.NoError(t, err)
	require.True(t, ok)

	require.NotZero(t, user.CreatedAt)
	require.NotZero(t, user.ID)

	return user
}

func TestUserModel_Insert(t *testing.T) {
	var user store.User

	t.Run("success case for insert user", func(t *testing.T) {
		user = createRandomUser(t)
	})

	t.Run("duplicate email case for insert user", func(t *testing.T) {
		err := testModels.Users.Insert(&user)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrDuplicateEmail)
	})
}

func TestUserModel_Get(t *testing.T) {
	user1 := createRandomUser(t)

	t.Run("success case for get user", func(t *testing.T) {
		user2, err := testModels.Users.Get(user1.ID)
		require.NoError(t, err)
		require.NotEmpty(t, user2)

		require.Equal(t, user1.Name, user2.Name)
		require.Equal(t, user1.Email, user2.Email)
		require.Equal(t, user1.Password.Hashed, user2.Password.Hashed)
		require.Equal(t, user1.IsActivated, user2.IsActivated)
		require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	})

	t.Run("not found case for get user", func(t *testing.T) {
		user2, err := testModels.Users.Get(user1.ID + 1)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)
		require.Empty(t, user2)
	})
}

func TestUserModel_GetByEmail(t *testing.T) {
	user1 := createRandomUser(t)

	t.Run("success case for get user by email", func(t *testing.T) {
		user2, err := testModels.Users.GetByEmail(user1.Email)
		require.NoError(t, err)
		require.NotEmpty(t, user2)

		require.Equal(t, user1.Name, user2.Name)
		require.Equal(t, user1.Email, user2.Email)
		require.Equal(t, user1.Password.Hashed, user2.Password.Hashed)
		require.Equal(t, user1.IsActivated, user2.IsActivated)
		require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
	})

	t.Run("not found case for get user by email", func(t *testing.T) {
		user2, err := testModels.Users.GetByEmail(random.Email())
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrRecordNotFound)
		require.Empty(t, user2)
	})
}

func TestUserModel_Update(t *testing.T) {
	user1 := createRandomUser(t)

	t.Run("success case for update user", func(t *testing.T) {
		user2 := user1
		user2.IsActivated = true

		err := testModels.Users.Update(&user2)
		require.NoError(t, err)
		require.NotEmpty(t, user2)

		require.Equal(t, user1.Name, user2.Name)
		require.Equal(t, user1.Email, user2.Email)
		require.Equal(t, user1.Password.Hashed, user2.Password.Hashed)
		require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)

		require.NotEqual(t, user1.IsActivated, user2.IsActivated)
		require.Equal(t, user1.Version+1, user2.Version)
	})

	t.Run("duplicate email case for update user", func(t *testing.T) {
		user3 := createRandomUser(t)

		user2 := user1
		user2.IsActivated = true
		user2.Version = user1.Version + 1
		user2.Email = user3.Email

		err := testModels.Users.Update(&user2)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrDuplicateEmail)
	})

	t.Run("edit conflict case for update user", func(t *testing.T) {
		user2 := user1
		user2.IsActivated = true

		err := testModels.Users.Update(&user2)
		require.Error(t, err)
		require.ErrorIs(t, err, store.ErrEditConflict)
	})

}

func TestUserModel_GetForToken(t *testing.T) {
	token := createNewToken(t)

	user, err := testModels.Users.GetForToken(token.Scope, token.Plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, user.ID, token.UserID)
}

func TestUserModel_GetAccounts(t *testing.T) {
	account := createRandomAccount(t)

	err := testModels.Accounts.AddUser(account.OwnerID, account.ID)
	require.NoError(t, err)

	accounts, err := testModels.Users.GetAccounts(account.OwnerID)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)

	require.Equal(t, account.ID, accounts[0].ID)
}
