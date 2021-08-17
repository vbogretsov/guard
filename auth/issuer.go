package auth

import (
	"fmt"
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
	refresh RefreshGenerator
	method  jwt.SigningMethod
}

func NewIssuer(secret string, timer Timer, ttl time.Duration, method jwt.SigningMethod, refresh RefreshGenerator) Issuer {
	return &issuer{
		secret:  []byte(secret),
		timer:   timer,
		ttl:     ttl,
		method:  method,
		refresh: refresh,
	}
}

func encodeJWT(secret []byte, claims map[string]interface{}, method jwt.SigningMethod) (string, error) {
	token := jwt.NewWithClaims(method, jwt.MapClaims(claims))
	return token.SignedString(secret)
}

func (c *issuer) Issue(user model.User) (Token, error) {
	var token Token

	refresh, err := c.refresh.Generate(user)
	if err != nil {
		return token, err
	}

	now := c.timer.Now()
	exp := now.Add(c.ttl).Unix()

	claims := map[string]interface{}{
		"sub": user.Name,
		"exp": exp,
		// TODO: add more claims.
	}

	access, err := encodeJWT(c.secret, claims, c.method)
	if err != nil {
		return token, fmt.Errorf("jwt encoding failed: %w", err)
	}

	token.IssuedAt = now.Unix()
	token.Access = string(access)
	token.AccessExpires = exp
	token.Refresh = refresh.ID
	token.RefreshExpires = refresh.Expires

	return token, nil
}
