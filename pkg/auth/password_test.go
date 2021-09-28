package auth_test

import (
	"github.com/nebisin/goExpense/pkg/auth"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPassword_Matches(t *testing.T) {
	plaintext := random.Password()
	pass := auth.Password{}

	err := pass.Set(plaintext)
	require.NoError(t, err)
	require.NotEmpty(t, pass)

	t.Run("success case for password matches", func(t *testing.T) {
		matches, err := pass.Matches(plaintext)
		require.NoError(t, err)
		require.True(t, matches)
	})

	t.Run("mismatched password for password matches", func(t *testing.T) {
		matches, err := pass.Matches(random.Password())
		require.NoError(t, err)
		require.False(t, matches)
	})
}
