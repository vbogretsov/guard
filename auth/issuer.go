package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/profile"
)

type IssuerConf struct {
	Key []byte
	TTL time.Duration
	Alg jwt.SigningMethod
}

type Issuer interface {
	Issue(user model.User) (Token, error)
}

type issuer struct {
	conf    IssuerConf
	timer   Timer
	refresh RefreshGenerator
	claimer profile.Claimer
}

func NewIssuer(conf IssuerConf, timer Timer, refresh RefreshGenerator, claimer profile.Claimer) Issuer {
	return &issuer{
		conf:    conf,
		timer:   timer,
		refresh: refresh,
		claimer: claimer,
	}
}

func encodeJWT(conf IssuerConf, claims map[string]interface{}) (string, error) {
	token := jwt.NewWithClaims(conf.Alg, jwt.MapClaims(claims))
	return token.SignedString(conf.Key)
}

func (c *issuer) Issue(user model.User) (Token, error) {
	var token Token

	refresh, err := c.refresh.Generate(user)
	if err != nil {
		return token, err
	}

	now := c.timer.Now()
	exp := now.Add(c.conf.TTL).Unix()

	cls, err := c.claimer.GetClaims(user.ID)
	if err != nil {
		return token, err
	}

	claims := map[string]interface{}{
		"sub": user.Name,
		"exp": exp,
	}

	for k, v := range cls {
		claims[k] = v
	}

	access, err := encodeJWT(c.conf, claims)
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
