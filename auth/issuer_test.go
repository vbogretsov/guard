package auth_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
)

type refreshGeneratorMock struct {
	mock.Mock
}

func (m *refreshGeneratorMock) Generate(user model.User) (model.RefreshToken, error) {
	args := m.Called(user)

	token := args.Get(0)
	if token == nil {
		return model.RefreshToken{}, args.Error(1)
	}

	return token.(model.RefreshToken), args.Error(1)
}

func decodeJWT(secret, access string) (*jwt.Token, error) {
	return jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}

type issuerMock struct {
	mock.Mock
}

func (m *issuerMock) Issue(user model.User) (auth.Token, error) {
	args := m.Called(user)

	token := args.Get(0)
	if token == nil {
		return auth.Token{}, args.Error(1)
	}

	return token.(auth.Token), args.Error(1)
}

type signingMethodMock struct {
	mock.Mock
}

func (m *signingMethodMock) Verify(signingString, signature string, key interface{}) error {
	panic("Verify not implemented")
}

func (m *signingMethodMock) Alg() string {
	return "mock"
}

func (m *signingMethodMock) Sign(signingString string, key interface{}) (string, error) {
	args := m.Called(signingString, key)
	return args.String(0), args.Error(1)
}

func TestIssuer(t *testing.T) {
	sm := jwt.SigningMethodHS256
	t.Run("Success", func(t *testing.T) {
		secret := "123.456"
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}

		accessTTL := 300 * time.Second
		refreshTTL := 3600 * time.Second

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		refreshToken := model.RefreshToken{
			UserID:  user.ID,
			User:    user,
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(refreshTTL).Unix(),
		}

		refresh.
			On("Generate", mock.MatchedBy(matchUser(user))).
			Return(refreshToken, nil)

		cmd := auth.NewIssuer(secret, timer, accessTTL, sm, refresh)

		token, err := cmd.Issue(user)
		require.NoError(t, err)

		expires := timer.Now().Add(accessTTL).Unix()

		require.Equal(t, timer.Now().Unix(), token.IssuedAt)
		require.Equal(t, expires, token.AccessExpires)
		require.Equal(t, refreshToken.Expires, token.RefreshExpires)

		raw, err := decodeJWT(secret, token.Access)
		require.NoError(t, err)
		require.Equal(t, user.Name, (raw.Claims).(jwt.MapClaims)["sub"])
		require.Equal(t, expires, int64((raw.Claims).(jwt.MapClaims)["exp"].(float64)))
	})

	t.Run("FailedCreateRefresh", func(t *testing.T) {
		secret := "123.456"
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}

		accessTTL := 300 * time.Second

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		fail := errors.New("xxx")

		refresh.
			On("Generate", mock.Anything).
			Return(nil, fail)

		cmd := auth.NewIssuer(secret, timer, accessTTL, sm, refresh)

		_, err := cmd.Issue(user)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailedJWTEncode", func(t *testing.T) {
		secret := ""
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}
		signing := &signingMethodMock{}

		accessTTL := 300 * time.Second
		refreshTTL := 3600 * time.Second

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		refreshToken := model.RefreshToken{
			UserID:  user.ID,
			User:    user,
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(refreshTTL).Unix(),
		}

		refresh.
			On("Generate", mock.MatchedBy(matchUser(user))).
			Return(refreshToken, nil)

		fail := errors.New("xxx")
		signing.On("Sign", mock.Anything, mock.Anything).Return("", fail)

		cmd := auth.NewIssuer(secret, timer, accessTTL, signing, refresh)

		_, err := cmd.Issue(user)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
