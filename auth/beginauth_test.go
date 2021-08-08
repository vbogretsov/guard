package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
)

func TestStartOAuth(t *testing.T) {
	ttl := 30 * time.Second
	timer := &timerMock{value: time.Now()}
	gSession := &sessionMock{}
	sessions := &sessionsMock{}
	provider := &providerMock{}

	session := model.Session{
		Value:   "beginauth.session.value",
		Created: timer.Now().Unix(),
		Expires: timer.Now().Add(ttl).Unix(),
	}

	authURL := "http://auth.url"

	gSession.On("Marshal").Return(session.Value)
	gSession.On("GetAuthURL").Return(authURL, nil)
	provider.On("BeginAuth", mock.Anything).Return(gSession, nil)
	sessions.On("Create", mock.MatchedBy(matchSession(session))).Return(nil)

	cmd := auth.NewOAuthStarter(ttl, timer, sessions, provider)

	result, err := cmd.StartOAuth()
	require.NoError(t, err)
	require.Equal(t, authURL, result)
}
