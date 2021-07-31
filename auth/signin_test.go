package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
)

func TestSignIn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tx := &txMock{}
		xsrf := &xsrfValidatorMock{}
		userFoC := &userFindOrCreatorMock{}
		issuer := &issuerMock{}

		xsrfToken := "signin.xsrf.123"

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

		tx.Default()
		xsrf.On("Validate", xsrfToken).Return(nil)
		userFoC.On("FindOrCreate", user.Name).Return(user, nil)
		issuer.On("Issue", user).Return(token, nil)

		cmd := auth.NewSignIner(tx, xsrf, userFoC, issuer)

		result, err := cmd.SignIn(xsrfToken, user.Name)
		require.NoError(t, err)
		require.Equal(t, token, result)
	})
}
