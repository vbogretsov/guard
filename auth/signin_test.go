package auth_test

import (
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

		issuer.On("Issue", user).Return(token, nil)

		cmd := auth.NewSignIner(validator, fetcher, issuer)

		result, err := cmd.SignIn(session.ID, nil)
		require.NoError(t, err)
		require.Equal(t, token, result)
	})
}
