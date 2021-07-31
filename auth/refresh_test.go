package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
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
	return args.Get(0).(model.RefreshToken), args.Error(1)
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

func TestRefreshTokenCreator(t *testing.T) {
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

		cmd := auth.NewRefreshTokenCreator(rtm, tm, ttl)

		result, err := cmd.Create(user)

		require.NoError(t, err)
		require.NotEmpty(t, result.ID)

		require.Equal(t, token.UserID, result.UserID)
		require.Equal(t, token.User, result.User)
		require.Equal(t, token.Created, result.Created)
		require.Equal(t, token.Expires, result.Expires)
	})
}

func TestRefreshToken(t *testing.T) {

}

// func TestRefreshTokenRetrieveCommand(t *testing.T) {
// 	ttl := 60 * time.Second

// 	user := model.User{
// 		ID:      "123",
// 		Name:    "u0@mail.com",
// 		Created: 1600000000,
// 	}

// 	t.Run("Fresh", func(t *testing.T) {
// 		rtm := &refreshTokensMock{}
// 		tm := &timerMock{value: time.Now()}

// 		token := model.RefreshToken{
// 			ID:      "123.123",
// 			UserID:  user.ID,
// 			User:    user,
// 			Created: tm.Now().Unix(),
// 			Expires: tm.Now().Add(ttl).Unix(),
// 		}

// 		rtm.On("Find", token.ID).Return(token, nil)
// 		rtm.On("Delete", token.ID)

// 		cmd := auth.NewRefreshTokenRetrieveCommand(rtm, tm)

// 		result, err := cmd.Execute(token.ID)
// 		require.NoError(t, err)
// 		require.Equal(t, token, result)
// 	})

// 	t.Run("Expired", func(t *testing.T) {
// 		rtm := &refreshTokensMock{}
// 		tm := &timerMock{value: time.Now()}

// 		token := model.RefreshToken{
// 			ID:      "123.123",
// 			UserID:  user.ID,
// 			User:    user,
// 			Created: tm.Now().Unix(),
// 			Expires: tm.Now().Add(ttl).Add(10 * time.Second).Unix(),
// 		}

// 		rtm.On("Find", token.ID).Return(token, nil)

// 		cmd := auth.NewRefreshTokenRetrieveCommand(rtm, tm)

// 		_, err := cmd.Execute(token.ID)
// 		require.Error(t, err)
// 		require.ErrorAs(t, err, &auth.Error{})
// 	})

// 	t.Run("Invalid", func(t *testing.T) {
// 		rtm := &refreshTokensMock{}
// 		tm := &timerMock{value: time.Now()}

// 		tokenID := "xxx"

// 		rtm.On("Find", tokenID).Return(model.RefreshToken{}, auth.Error{})

// 		cmd := auth.NewRefreshTokenRetrieveCommand(rtm, tm)

// 		_, err := cmd.Execute(tokenID)
// 		require.Error(t, err)
// 		require.ErrorAs(t, err, &auth.Error{})
// 	})
// }
