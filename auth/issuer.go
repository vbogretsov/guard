package auth

import (
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/vbogretsov/guard/model"
)

type Issuer interface {
	Issue(user model.User) (Token, error)
}

type issuer struct {
	secret  []byte
	timer   Timer
	ttl     time.Duration
	refresh RefreshTokenCreator
}

func NewIssuer(secret string, timer Timer, ttl time.Duration, refresh RefreshTokenCreator) Issuer {
	return &issuer{
		secret:  []byte(secret),
		timer:   timer,
		ttl:     ttl,
		refresh: refresh,
	}
}

func encodeJWT(secret []byte, claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims(claims))
	return token.SignedString(secret)
}

func (c *issuer) Issue(user model.User) (Token, error) {
	var token Token

	refresh, err := c.refresh.Create(user)
	if err != nil {
		return token, err
	}

	now := c.timer.Now()
	exp := now.Add(c.ttl).Unix()

	access, err := encodeJWT(c.secret, map[string]interface{}{
		"sub": user.Name,
		"exp": exp,
		// TODO: add more claims.
	})

	if err != nil {
		return token, err
	}

	token.IssuedAt = now.Unix()
	token.Access = string(access)
	token.AccessExpires = exp
	token.Refresh = refresh.ID
	token.RefreshExpires = refresh.Expires

	return token, nil
}
