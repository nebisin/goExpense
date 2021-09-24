package store_test

import (
	"github.com/nebisin/goExpense/internal/store"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createNewToken(t *testing.T) *store.Token {
	user := createRandomUser(t)

	token, err := testModels.Tokens.New(user.ID, time.Hour, store.ScopeActivation)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	require.Equal(t, token.UserID, user.ID)
	require.Equal(t, token.Scope, store.ScopeActivation)
	require.WithinDuration(t, token.Expiry, time.Now().Add(time.Hour), time.Second)

	return token
}

func TestTokenModel_New(t *testing.T) {
	createNewToken(t)
}

func TestTokenModel_DeleteAllForUser(t *testing.T) {
	token := createNewToken(t)

	err := testModels.Tokens.DeleteAllForUser(token.Scope, token.UserID)
	require.NoError(t, err)

	user, err := testModels.Users.GetForToken(token.Scope, token.Plaintext)
	require.Error(t, err)
	require.ErrorIs(t, err, store.ErrRecordNotFound)
	require.Empty(t, user)
}