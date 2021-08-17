package auth_test

import (
	"testing"
	"time"

	"github.com/markbates/goth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"

	"github.com/vbogretsov/guard/auth"
)

type timerMock struct {
	value time.Time
}

func (m *timerMock) Now() time.Time {
	return m.value
}

type sessionMock struct {
	mock.Mock
}

func (m *sessionMock) Authorize(provider goth.Provider, params goth.Params) (string, error) {
	args := m.Called(provider, params)

	result := args.Get(0)
	if result == nil {
		return "", args.Error(1)
	}

	return result.(string), args.Error(1)
}

func (m *sessionMock) GetAuthURL() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *sessionMock) Marshal() string {
	return m.Called().String(0)
}

type providerMock struct {
	mock.Mock
}

func (m *providerMock) UnmarshalSession(raw string) (goth.Session, error) {
	args := m.Called(raw)

	sess := args.Get(0)
	if sess == nil {
		return nil, args.Error(1)
	}

	return sess.(goth.Session), args.Error(1)
}

func (m *providerMock) FetchUser(sess goth.Session) (goth.User, error) {
	args := m.Called(sess)

	user := args.Get(0)
	if user == nil {
		return goth.User{}, args.Error(1)
	}

	return user.(goth.User), args.Error(1)
}

func (m *providerMock) BeginAuth(state string) (goth.Session, error) {
	args := m.Called(state)

	sess := args.Get(0)
	if sess == nil {
		return nil, args.Error(1)
	}

	return sess.(goth.Session), args.Error(1)
}

func (m *providerMock) Name() string {
	return m.Called().String(0)
}

func (m *providerMock) SetName(name string) {
	m.Called(name)
}

func (m *providerMock) Debug(debug bool) {
	m.Called(debug)
}

func (m *providerMock) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	args := m.Called(refreshToken)

	token := args.Get(0)
	if token == nil {
		return nil, args.Error(1)
	}

	return token.(*oauth2.Token), args.Error(1)
}

func (m *providerMock) RefreshTokenAvailable() bool {
	return m.Called().Get(0).(bool)
}

func TestTimer(t *testing.T) {
	tm := auth.RealTimer{}

	now := time.Now()

	v1 := tm.Now()
	require.GreaterOrEqual(t, now.Unix(), v1.Unix())

	v2 := tm.Now()
	require.Equal(t, v1, v2)
}
