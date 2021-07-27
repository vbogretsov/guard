package auth

import "time"

type Error struct {
	Msg string
}

func (e Error) Error() string {
	return e.Msg
}

type Token struct {
	IssuedAt   int64
	Access     string
	AccessTTL  int64
	Refresh    string
	RefreshTTL string
}

type Timer interface {
	Now() time.Time
}
