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

type xsrfMock struct {
	mock.Mock
}

func (m *xsrfMock) Find(id string) (model.XSRFToken, error) {
	args := m.Called(id)

	value := args.Get(0)
	if value == nil {
		return model.XSRFToken{}, args.Error(1)
	}

	return value.(model.XSRFToken), args.Error(1)
}

func (m *xsrfMock) Create(token model.XSRFToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *xsrfMock) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func matchXSRF(token model.XSRFToken) func(model.XSRFToken) bool {
	return func(arg model.XSRFToken) bool {
		return token.Created == arg.Created &&
			token.Expires == arg.Expires
	}
}

type xsrfGeneratorMock struct {
	mock.Mock
}

func (m *xsrfGeneratorMock) Generate() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type xsrfValidatorMock struct {
	mock.Mock
}

func (m *xsrfValidatorMock) Validate(value string) error {
	args := m.Called(value)
	return args.Error(0)
}

func TestXSRFGenerator(t *testing.T) {
	tokens := &xsrfMock{}
	timer := &timerMock{value: time.Now()}
	ttl := 3600 * time.Second

	value := model.XSRFToken{
		Created: timer.Now().Unix(),
		Expires: timer.Now().Add(ttl).Unix(),
	}

	tokens.
		On("Create", mock.MatchedBy(matchXSRF(value))).
		Return(nil)

	cmd := auth.NewXSRFGenerator(tokens, timer, ttl)

	id, err := cmd.Generate()
	require.NoError(t, err)
	require.NotEmpty(t, id)
}

func TestXSRFValidator(t *testing.T) {
	t.Run("Fresh", func(t *testing.T) {
		tokens := &xsrfMock{}
		timer := &timerMock{value: time.Now()}

		value := model.XSRFToken{
			ID:      "xsrf.123",
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		tokens.On("Find", value.ID).Return(value, nil)
		tokens.On("Delete", value.ID).Return(nil)

		cmd := auth.NewXSRFValidator(tokens, timer)

		err := cmd.Validate(value.ID)
		require.NoError(t, err)
	})

	t.Run("Expired", func(t *testing.T) {
		tokens := &xsrfMock{}
		timer := &timerMock{value: time.Now()}

		value := model.XSRFToken{
			ID:      "xsrf.123",
			Created: timer.Now().Unix(),
			Expires: timer.Now().Add(3600 * time.Second).Unix(),
		}

		timer.value = time.Now().Add(4600 * time.Second)

		tokens.On("Find", value.ID).Return(value, nil)
		tokens.On("Delete", value.ID).Return(nil)

		cmd := auth.NewXSRFValidator(tokens, timer)

		err := cmd.Validate(value.ID)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})

	t.Run("Invalid", func(t *testing.T) {
		tokens := &xsrfMock{}
		timer := &timerMock{value: time.Now()}

		value := "xxx"

		tokens.On("Find", value).Return(nil, repo.ErrorNotFound)

		cmd := auth.NewXSRFValidator(tokens, timer)

		err := cmd.Validate(value)
		require.Error(t, err)
		require.ErrorAs(t, err, &auth.Error{})
	})
}
