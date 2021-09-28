package auth_test

import (
	"github.com/nebisin/goExpense/pkg/auth"
	"github.com/nebisin/goExpense/pkg/config"
	"github.com/nebisin/goExpense/pkg/random"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
	"time"
)

func TestJWTMaker_VerifyToken(t *testing.T) {
	cfg, err := config.LoadConfig("../..", "test")
	if err != nil {
		log.Fatal(err)
	}

	maker, err := auth.NewJWTMaker(cfg.JwtSecret)
	require.NoError(t, err)
	require.NotEmpty(t, maker)

	t.Run("success case for verify token", func(t *testing.T) {
		userID := random.Int(2, 6)

		token, err := maker.CreateToken(userID, time.Hour)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotEmpty(t, payload)

		require.Equal(t, payload.UserID, userID)
	})

	t.Run("expired token case for verify token", func(t *testing.T) {
		userID := random.Int(2, 6)

		token, err := maker.CreateToken(userID, time.Nanosecond-2)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payload, err := maker.VerifyToken(token)
		require.Error(t, err)
		require.ErrorIs(t, err, auth.ErrExpiredToken)
		require.Empty(t, payload)
	})

	t.Run("invalid token case for verify token", func(t *testing.T) {
		payload, err := maker.VerifyToken("token")
		require.Error(t, err)
		require.ErrorIs(t, err, auth.ErrInvalidToken)
		require.Empty(t, payload)
	})
}
