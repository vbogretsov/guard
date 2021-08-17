package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

type sessionsMock struct {
	mock.Mock
}

func (m *sessionsMock) Find(id string) (model.Session, error) {
	args := m.Called(id)

	value := args.Get(0)
	if value == nil {
		return model.Session{}, args.Error(1)
	}

	return value.(model.Session), args.Error(1)
}

func (m *sessionsMock) Create(value model.Session) error {
	args := m.Called(value)
	return args.Error(0)
}

func (m *sessionsMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func matchSession(sess model.Session) func(model.Session) bool {
	return func(arg model.Session) bool {
		return sess.Created == arg.Created &&
			sess.Expires == arg.Expires &&
			sess.Value == arg.Value
	}
}

func TestSignIn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		sessions := &sessionsMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		session := model.Session{
			ID:      "singin.session.id.123",
			Value:   "signin.session.value.123",
			Created: 1600000000,
			Expires: 1600000100,
		}

		user := model.User{
			ID:      "signin.user.123",
			Name:    "u0@mial.org",
			Created: 1600000000,
		}

		token := auth.Token{
			IssuedAt:       1640000000,
			Access:         "signin.access.123",
			AccessExpires:  1640000010,
			Refresh:        "signin.refresh.123",
			RefreshExpires: 1640000020,
		}

		sessions.On("Find", session.ID).Return(session, nil)
		fetcher.On("Fetch", session.Value, nil).Return(user, nil)
		issuer.On("Issue", user).Return(token, nil)

		cmd := auth.NewSignIner(sessions, fetcher, issuer)

		result, err := cmd.SignIn(session.ID, nil)
		require.NoError(t, err)
		require.Equal(t, token, result)
	})

	t.Run("FailOnSessionFind", func(t *testing.T) {
		sessions := &sessionsMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		sessionID := "singin.session.id.123"
		fail := errors.New("unexpected error")

		sessions.On("Find", sessionID).Return(nil, fail)

		cmd := auth.NewSignIner(sessions, fetcher, issuer)

		_, err := cmd.SignIn(sessionID, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("SessionNotFound", func(t *testing.T) {
		sessions := &sessionsMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		sessionID := "singin.session.id.123"

		sessions.On("Find", sessionID).Return(nil, repo.ErrorNotFound)

		cmd := auth.NewSignIner(sessions, fetcher, issuer)

		_, err := cmd.SignIn(sessionID, nil)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})

	t.Run("FailOnFetch", func(t *testing.T) {
		sessions := &sessionsMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		session := model.Session{
			ID:      "singin.session.id.123",
			Value:   "signin.session.value.123",
			Created: 1600000000,
			Expires: 1600000100,
		}

		fail := errors.New("xxx")

		sessions.On("Find", session.ID).Return(session, nil)
		fetcher.On("Fetch", session.Value, nil).Return(nil, fail)

		cmd := auth.NewSignIner(sessions, fetcher, issuer)

		_, err := cmd.SignIn(session.ID, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnIssue", func(t *testing.T) {
		sessions := &sessionsMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		session := model.Session{
			ID:      "singin.session.id.123",
			Value:   "signin.session.value.123",
			Created: 1600000000,
			Expires: 1600000100,
		}

		user := model.User{
			ID:      "signin.user.123",
			Name:    "u0@mial.org",
			Created: 1600000000,
		}

		fail := errors.New("xxx")

		sessions.On("Find", session.ID).Return(session, nil)
		fetcher.On("Fetch", session.Value, nil).Return(user, nil)
		issuer.On("Issue", user).Return(nil, fail)

		cmd := auth.NewSignIner(sessions, fetcher, issuer)

		_, err := cmd.SignIn(session.ID, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
