package auth

import (
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/markbates/goth"
)

var signingMethod = jwt.SigningMethodHS256

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}

// TODO: use oauth2.Token
type Token struct {
	IssuedAt       int64
	Access         string
	AccessExpires  int64
	Refresh        string
	RefreshExpires int64
}

type Timer interface {
	Now() time.Time
}

type RealTimer struct {
	value time.Time
}

func (rt *RealTimer) Now() time.Time {
	var zero time.Time

	if rt.value == zero {
		rt.value = time.Now()
	}

	return rt.value
}

type Factory interface {
	NewOAuthStarter(provider goth.Provider) OAuthStarter
	NewSignIner(provider goth.Provider) SignIner
	NewRefresher() Refresher
}
