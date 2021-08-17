package auth_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

type refreshTokensMock struct {
	mock.Mock
}

func (m *refreshTokensMock) Create(token model.RefreshToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *refreshTokensMock) Find(id string) (model.RefreshToken, error) {
	args := m.Called(id)

	token := args.Get(0)
	if token == nil {
		return model.RefreshToken{}, args.Error(1)
	}

	return token.(model.RefreshToken), args.Error(1)
}

func (m *refreshTokensMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func matchRefreshToken(token model.RefreshToken) func(model.RefreshToken) bool {
	return func(arg model.RefreshToken) bool {
		return token.UserID == arg.UserID &&
			token.Created == arg.Created &&
			token.Expires == arg.Expires
	}
}

func TestRefreshGenerator(t *testing.T) {
	ttl := 60 * time.Second

	user := model.User{
		ID:      "123",
		Name:    "u0@mail.com",
		Created: 1600000000,
	}

	t.Run("Success", func(t *testing.T) {
		rtm := &refreshTokensMock{}
		tm := &timerMock{value: time.Now()}

		token := model.RefreshToken{
			UserID:  user.ID,
			User:    user,
			Created: tm.Now().Unix(),
			Expires: tm.Now().Add(ttl).Unix(),
		}

		rtm.On("Create", mock.MatchedBy(matchRefreshToken(token))).Return(nil)

		cmd := auth.NewRefreshGenerator(rtm, tm, ttl)

		result, err := cmd.Generate(user)

		require.NoError(t, err)
		require.NotEmpty(t, result.ID)

		require.Equal(t, token.UserID, result.UserID)
		require.Equal(t, token.User, result.User)
		require.Equal(t, token.Created, result.Created)
		require.Equal(t, token.Expires, result.Expires)
	})

	t.Run("Failed", func(t *testing.T) {
		rtm := &refreshTokensMock{}
		tm := &timerMock{value: time.Now()}

		fail := errors.New("xxx")

		rtm.On("Create", mock.Anything).Return(fail)

		cmd := auth.NewRefreshGenerator(rtm, tm, ttl)

		_, err := cmd.Generate(user)

		require.ErrorIs(t, err, fail)
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("Fresh", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		tokens := &refreshTokensMock{}
		issuer := &issuerMock{}

		user := model.User{
			ID:      "refresh.123",
			Name:    "u0@mail.com",
			Created: 1600000000,
		}

		refresh := model.RefreshToken{
			ID:      "refresh.123",
			UserID:  user.ID,
			User:    user,
			Created: time.Now().Unix(),
			Expires: time.Now().Add(3600 * time.Second).Unix(),
		}

		timer.value = time.Now().Add(2600 * time.Second)

		tokens.On("Find", refresh.ID).Return(refresh, nil)
		tokens.On("Delete", refresh.ID).Return(nil)
		issuer.On("Issue", user).Return(auth.Token{}, nil)

		cmd := auth.NewRefresher(timer, tokens, issuer)

		_, err := cmd.Refresh(refresh.ID)
		require.NoError(t, err)
	})

	t.Run("Expired", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		tokens := &refreshTokensMock{}

		refresh := model.RefreshToken{
			ID:      "refresh.123",
			UserID:  "xxx",
			Created: time.Now().Unix(),
			Expires: time.Now().Add(3600 * time.Second).Unix(),
		}

		timer.value = time.Now().Add(4600 * time.Second)

		tokens.On("Find", refresh.ID).Return(refresh, nil)

		cmd := auth.NewRefresher(timer, tokens, &issuerMock{})

		_, err := cmd.Refresh(refresh.ID)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})

	t.Run("Invalid", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		tokens := &refreshTokensMock{}

		timer.value = time.Now().Add(4600 * time.Second)

		refreshToken := "xxx"

		tokens.On("Find", refreshToken).Return(nil, repo.ErrorNotFound)

		cmd := auth.NewRefresher(timer, tokens, &issuerMock{})

		_, err := cmd.Refresh(refreshToken)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})

	t.Run("FainOldFailed", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		tokens := &refreshTokensMock{}

		timer.value = time.Now().Add(4600 * time.Second)

		refreshToken := "xxx"
		fail := errors.New("xxx")

		tokens.On("Find", refreshToken).Return(nil, fail)

		cmd := auth.NewRefresher(timer, tokens, &issuerMock{})

		_, err := cmd.Refresh(refreshToken)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("IssueNewFailed", func(t *testing.T) {
		timer := &timerMock{value: time.Now()}
		issuer := &issuerMock{}
		tokens := &refreshTokensMock{}

		refresh := model.RefreshToken{
			ID:      "refresh.123",
			UserID:  "xxx",
			Created: time.Now().Unix(),
			Expires: time.Now().Add(3600 * time.Second).Unix(),
		}

		timer.value = time.Now().Add(2600 * time.Second)

		fail := errors.New("xxx")

		tokens.On("Find", refresh.ID).Return(refresh, nil)
		issuer.On("Issue", mock.Anything).Return(nil, fail)

		cmd := auth.NewRefresher(timer, tokens, issuer)

		_, err := cmd.Refresh(refresh.ID)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
