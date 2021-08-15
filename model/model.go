package model

type User struct {
	ID      string
	Name    string
	Created int64
}

type RefreshToken struct {
	ID      string
	UserID  string
	User    User
	Created int64
	Expires int64
}

type Session struct {
	ID      string
	Value   string
	Created int64
	Expires int64
}
