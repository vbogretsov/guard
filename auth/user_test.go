package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/auth"
	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

type usersMock struct {
	mock.Mock
}

func (m *usersMock) Create(user model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *usersMock) Find(name string) (model.User, error) {
	args := m.Called(name)

	user := args.Get(0)
	if user == nil {
		return model.User{}, args.Error(1)
	}

	return user.(model.User), args.Error(1)
}

func matchUser(user model.User) func(model.User) bool {
	return func(arg model.User) bool {
		return user.Name == arg.Name &&
			user.Created == arg.Created
	}
}

func TestUserProvideCommand(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		um := &usersMock{}
		tm := &timerMock{value: time.Now()}

		user := model.User{
			Name:    "u0@mail.org",
			Created: tm.Now().Unix(),
		}

		um.On("Find", user.Name).
			Return(model.User{}, repo.ErrorNotFound)
		um.On("Create", mock.MatchedBy(matchUser(user))).
			Return(nil)

		cmd := auth.NewUserProvideCommand(um, tm)

		result, err := cmd.Execute(user.Name)
		require.NoError(t, err)
		require.NotEmpty(t, result.ID)
	})

	t.Run("Old", func(t *testing.T) {
		um := &usersMock{}
		tm := &timerMock{value: time.Now()}

		user := model.User{
			Name:    "u0@mail.org",
			Created: tm.Now().Unix(),
		}

		um.On("Find", user.Name).Return(user, nil)

		cmd := auth.NewUserProvideCommand(um, tm)

		result, err := cmd.Execute(user.Name)
		require.NoError(t, err)
		require.Equal(t, user, result)
	})
}
