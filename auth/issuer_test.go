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

type claimerMock struct {
	mock.Mock
}

func (m *claimerMock) GetClaims(userID string) (map[string]interface{}, error) {
	args := m.Called(userID)

	claims := args.Get(0)
	if claims == nil {
		return nil, args.Error(1)
	}

	return claims.(map[string]interface{}), args.Error(1)
}

func decodeJWT(secret []byte, access string) (*jwt.Token, error) {
	return jwt.Parse(access, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return secret, nil
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
	conf := auth.IssuerConf{
		Key: []byte("123456"),
		TTL: 300 * time.Second,
		Alg: jwt.SigningMethodHS256,
	}

	t.Run("Success", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}
		claimer := &claimerMock{}

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		refreshToken := model.RefreshToken{
			UserID:  user.ID,
			User:    user,
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		claims := map[string]interface{}{
			"claims": map[string]interface{}{
				"x-key1": "x-val1",
				"x-key2": "x-val2",
			},
		}

		refresh.
			On("Generate", mock.MatchedBy(matchUser(user))).
			Return(refreshToken, nil)

		claimer.
			On("GetClaims", user.ID).
			Return(claims, nil)

		cmd := auth.NewIssuer(conf, timer, refresh, claimer)

		token, err := cmd.Issue(user)
		require.NoError(t, err)

		expires := timer.Now().Add(conf.TTL).Unix()

		require.Equal(t, timer.Now().Unix(), token.IssuedAt)
		require.Equal(t, expires, token.AccessExpires)
		require.Equal(t, refreshToken.Expires, token.RefreshExpires)

		raw, err := decodeJWT(conf.Key, token.Access)
		require.NoError(t, err)
		require.Equal(t, user.Name, (raw.Claims).(jwt.MapClaims)["sub"])
		require.Equal(t, expires, int64((raw.Claims).(jwt.MapClaims)["exp"].(float64)))
		require.Equal(t, claims["claims"], (raw.Claims).(jwt.MapClaims)["claims"].(map[string]interface{}))
	})

	t.Run("FailedCreateRefresh", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}
		claimer := &claimerMock{}

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		fail := errors.New("xxx")

		refresh.
			On("Generate", mock.Anything).
			Return(nil, fail)

		cmd := auth.NewIssuer(conf, timer, refresh, claimer)

		_, err := cmd.Issue(user)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("ClaimerFailed", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}
		claimer := &claimerMock{}

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		refreshToken := model.RefreshToken{
			UserID:  user.ID,
			User:    user,
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		refresh.
			On("Generate", mock.MatchedBy(matchUser(user))).
			Return(refreshToken, nil)

		fail := errors.New("xxx")
		claimer.On("GetClaims", user.ID).Return(nil, fail)

		cmd := auth.NewIssuer(conf, timer, refresh, claimer)

		_, err := cmd.Issue(user)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("FailedJWTEncode", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		refresh := &refreshGeneratorMock{}
		signing := &signingMethodMock{}
		claimer := &claimerMock{}

		user := model.User{
			ID:      "issuer.user.123",
			Name:    "u0@mail.org",
			Created: timer.Now().Unix(),
		}

		refreshToken := model.RefreshToken{
			UserID:  user.ID,
			User:    user,
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		refresh.
			On("Generate", mock.MatchedBy(matchUser(user))).
			Return(refreshToken, nil)

		claims := map[string]interface{}{
			"claims": map[string]interface{}{
				"x-key1": "x-val1",
				"x-key2": "x-val2",
			},
		}

		claimer.
			On("GetClaims", user.ID).
			Return(claims, nil)

		fail := errors.New("xxx")
		signing.On("Sign", mock.Anything, mock.Anything).Return("", fail)

		confCpy := conf
		confCpy.Alg = signing

		cmd := auth.NewIssuer(confCpy, timer, refresh, claimer)

		_, err := cmd.Issue(user)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
