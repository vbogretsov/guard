package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
)

func TestStartOAuth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
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
	})

	t.Run("BeginAuthFailed", func(t *testing.T) {
		ttl := 30 * time.Second
		timer := &timerMock{value: time.Now()}
		// gSession := &sessionMock{}
		sessions := &sessionsMock{}
		provider := &providerMock{}

		// session := model.Session{
		// 	Value:   "beginauth.session.value",
		// 	Created: timer.Now().Unix(),
		// 	Expires: timer.Now().Add(ttl).Unix(),
		// }

		// authURL := "http://auth.url"
		fail := errors.New("xxx")

		provider.On("BeginAuth", mock.Anything).Return(nil, fail)
		// gSession.On("Marshal").Return(session.Value)
		// gSession.On("GetAuthURL").Return(authURL, nil)
		// sessions.On("Create", mock.MatchedBy(matchSession(session))).Return(nil)

		cmd := auth.NewOAuthStarter(ttl, timer, sessions, provider)

		_, err := cmd.StartOAuth()
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("CreateFailed", func(t *testing.T) {
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

		// authURL := "http://auth.url"
		fail := errors.New("xxx")

		provider.On("BeginAuth", mock.Anything).Return(gSession, nil)
		gSession.On("Marshal").Return(session.Value)
		sessions.On("Create", mock.Anything).Return(fail)
		// gSession.On("GetAuthURL").Return(authURL, nil)

		cmd := auth.NewOAuthStarter(ttl, timer, sessions, provider)

		_, err := cmd.StartOAuth()
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("GetAuthURL", func(t *testing.T) {
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

		fail := errors.New("xxx")

		provider.On("BeginAuth", mock.Anything).Return(gSession, nil)
		gSession.On("Marshal").Return(session.Value)
		sessions.On("Create", mock.Anything).Return(nil)
		gSession.On("GetAuthURL").Return("", fail)

		cmd := auth.NewOAuthStarter(ttl, timer, sessions, provider)

		_, err := cmd.StartOAuth()
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
