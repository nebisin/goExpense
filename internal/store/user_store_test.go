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

func TestInsertUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testModels.Users.Get(user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Password.Hashed, user2.Password.Hashed)
	require.Equal(t, user1.IsActivated, user2.IsActivated)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}

func TestGetUserByEmail(t *testing.T) {
	user1 := createRandomUser(t)

	user2, err := testModels.Users.GetByEmail(user1.Email)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.Password.Hashed, user2.Password.Hashed)
	require.Equal(t, user1.IsActivated, user2.IsActivated)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)
}
