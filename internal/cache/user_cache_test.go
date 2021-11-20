package cache_test

import (
	"testing"
	"time"

	"github.com/nebisin/goExpense/internal/store"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
)

func TestUserCache(t *testing.T) {
	randomName := random.Name()
	randomEmail := random.Email()
	randomPassword := random.Password()

	user1 := &store.User{
		Name:        randomName,
		Email:       randomEmail,
		IsActivated: false,
	}

	err := user1.Password.Set(randomPassword)
	require.NoError(t, err)

	err = testCache.User.Set(user1)
	require.NoError(t, err)

	user2, err := testCache.User.Get(user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.IsActivated, user2.IsActivated)
	require.WithinDuration(t, user1.CreatedAt, user2.CreatedAt, time.Second)

}
