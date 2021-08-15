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

type sessionValidatorMock struct {
	mock.Mock
}

func (m *sessionValidatorMock) Validate(id string) (model.Session, error) {
	args := m.Called(id)

	session := args.Get(0)
	if session == nil {
		return model.Session{}, args.Error(1)
	}

	return session.(model.Session), args.Error(1)
}

func TestSessionValidator(t *testing.T) {
	t.Run("Fresh", func(t *testing.T) {
		sessions := &sessionsMock{}
		timer := &timerMock{value: time.Now()}

		value := model.Session{
			ID:      "sessions.123",
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		sessions.On("Find", value.ID).Return(value, nil)
		sessions.On("Delete", value.ID).Return(nil)

		cmd := auth.NewSessionValidator(sessions, timer)

		result, err := cmd.Validate(value.ID)
		require.NoError(t, err)
		require.Equal(t, value, result)
	})

	t.Run("Expired", func(t *testing.T) {
		sessions := &sessionsMock{}
		timer := &timerMock{value: time.Now()}

		value := model.Session{
			ID:      "session.id.123",
			Value:   "session.value.123",
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		timer.value = time.Now().Add(4600 * time.Second)

		sessions.On("Find", value.ID).Return(value, nil)
		sessions.On("Delete", value.ID).Return(nil)

		cmd := auth.NewSessionValidator(sessions, timer)

		_, err := cmd.Validate(value.ID)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})

	t.Run("Invalid", func(t *testing.T) {
		sessions := &sessionsMock{}
		timer := &timerMock{value: time.Now()}

		value := "xxx"

		sessions.On("Find", value).Return(nil, repo.ErrorNotFound)

		cmd := auth.NewSessionValidator(sessions, timer)

		_, err := cmd.Validate(value)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})

	t.Run("Failed", func(t *testing.T) {
		sessions := &sessionsMock{}
		timer := &timerMock{value: time.Now()}

		value := "xxx"
		fail := errors.New("xxx")

		sessions.On("Find", value).Return(nil, fail)

		cmd := auth.NewSessionValidator(sessions, timer)

		_, err := cmd.Validate(value)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
