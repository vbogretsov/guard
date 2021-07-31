package auth_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/guard/auth"
)

type timerMock struct {
	value time.Time
}

func (m *timerMock) Now() time.Time {
	return m.value
}

type txMock struct {
	mock.Mock
}

func (m *txMock) Begin() error {
	args := m.Called()
	return args.Error(0)
}

func (m *txMock) Commit() error {
	args := m.Called()
	return args.Error(0)
}

func (m *txMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *txMock) Default() {
	m.On("Begin").Return(nil)
	m.On("Commit").Return(nil)
	m.On("Close").Return(nil)
}

func TestTimer(t *testing.T) {
	tm := auth.RealTimer{}

	now := time.Now()

	v1 := tm.Now()
	require.GreaterOrEqual(t, now.Unix(), v1.Unix())

	v2 := tm.Now()
	require.Equal(t, v1, v2)
}
