package auth_test

import "time"

type timerMock struct {
	value time.Time
}

func (m *timerMock) Now() time.Time {
	return m.value
}
