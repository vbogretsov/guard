package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
)

func TestStartOAuth(t *testing.T) {
	ttl := 30 * time.Second
	timer := &timerMock{}
	sessions := &sessionsMock{}
	provider := &providerMock{}

	cmd := auth.NewOAuthStarter(ttl, timer, sessions, provider)

	result, err := cmd.StartOAuth()
	require.NoError(t, err)
	require.NotEmpty(t, result)
}
