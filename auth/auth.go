package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/vbogretsov/guard/repo"
)

var signingMethod = jwt.SigningMethodHS256

type Error struct {
	msg string
}

func (e Error) Error() string {
	return e.msg
}

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

type OAuthStarter interface {
	StartOAuth(providerName string) (string, error)
}

type oauthStarter struct {
	xsrf XSRFGenerator
}

func NewOAuthStarter(xsrf XSRFGenerator) OAuthStarter {
	return &oauthStarter{
		xsrf: xsrf,
	}
}

func (c *oauthStarter) StartOAuth(providerName string) (string, error) {
	return "", errors.New("not implemented")
}

type SignIner interface {
	SignIn(xsrfToken, username string) (Token, error)
}

type signiner struct {
	tx     repo.Transaction
	xsrf   XSRGValidator
	user   UserFindOrCreator
	issuer Issuer
}

func NewSignIner(tx repo.Transaction, xsrf XSRGValidator, user UserFindOrCreator, issuer Issuer) SignIner {
	return &signiner{
		xsrf:   xsrf,
		user:   user,
		issuer: issuer,
	}
}

func (c *signiner) SignIn(xsrfToken, username string) (Token, error) {
	var empty Token

	if err := c.tx.Begin(); err != nil {
		return empty, err
	}
	defer c.tx.Close()

	if err := c.xsrf.Validate(xsrfToken); err != nil {
		return empty, err
	}

	user, err := c.user.FindOrCreate(username)
	if err != nil {
		return empty, err
	}

	token, err := c.issuer.Issue(user)
	if err != nil {
		return empty, err
	}

	if err := c.tx.Commit(); err != nil {
		return empty, err
	}

	return token, nil
}

type Factory interface {
	NewOAuthStarter() OAuthStarter
	NewSignIner() SignIner
	NewRefresher() Refresher
}
