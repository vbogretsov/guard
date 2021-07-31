package auth_test

import (
	"time"

	"github.com/stretchr/testify/mock"
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
