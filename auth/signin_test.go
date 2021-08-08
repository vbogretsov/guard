package auth_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
)

func TestSignIn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		validator := &sessionValidatorMock{}
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

		validator.On("Validate", session.ID).Return(session, nil)
		fetcher.On("Fetch", session.Value, nil).Return(user, nil)
		issuer.On("Issue", user).Return(token, nil)

		cmd := auth.NewSignIner(validator, fetcher, issuer)

		result, err := cmd.SignIn(session.ID, nil)
		require.NoError(t, err)
		require.Equal(t, token, result)
	})

	t.Run("FailOnValidate", func(t *testing.T) {
		validator := &sessionValidatorMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		sessionID := "singin.session.id.123"
		fail := errors.New("xxx")

		validator.On("Validate", sessionID).Return(nil, fail)

		cmd := auth.NewSignIner(validator, fetcher, issuer)

		_, err := cmd.SignIn(sessionID, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnFetch", func(t *testing.T) {
		validator := &sessionValidatorMock{}
		fetcher := &userFetcherMock{}
		issuer := &issuerMock{}

		session := model.Session{
			ID:      "singin.session.id.123",
			Value:   "signin.session.value.123",
			Created: 1600000000,
			Expires: 1600000100,
		}

		fail := errors.New("xxx")

		validator.On("Validate", session.ID).Return(session, nil)
		fetcher.On("Fetch", session.Value, nil).Return(nil, fail)

		cmd := auth.NewSignIner(validator, fetcher, issuer)

		_, err := cmd.SignIn(session.ID, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailOnIssue", func(t *testing.T) {
		validator := &sessionValidatorMock{}
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

		validator.On("Validate", session.ID).Return(session, nil)
		fetcher.On("Fetch", session.Value, nil).Return(user, nil)
		issuer.On("Issue", user).Return(nil, fail)

		cmd := auth.NewSignIner(validator, fetcher, issuer)

		_, err := cmd.SignIn(session.ID, nil)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
