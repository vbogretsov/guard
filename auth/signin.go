package auth

import (
	"fmt"

	"github.com/markbates/goth"
)

type SignIner interface {
	SignIn(code string, params goth.Params) (Token, error)
}

type signiner struct {
	validator SessionValidator
	fetcher   UserFetcher
	issuer    Issuer
}

func NewSignIner(validator SessionValidator, fetcher UserFetcher, issuer Issuer) SignIner {
	return &signiner{
		validator: validator,
		fetcher:   fetcher,
		issuer:    issuer,
	}
}

func (c *signiner) SignIn(state string, params goth.Params) (Token, error) {
	var empty Token

	session, err := c.validator.Validate(state)
	if err != nil {
		return empty, fmt.Errorf("session validation failed: %w", err)
	}

	user, err := c.fetcher.Fetch(session.Value, params)
	if err != nil {
		return empty, fmt.Errorf("fetch user failed: %w", err)
	}

	token, err := c.issuer.Issue(user)
	if err != nil {
		return empty, fmt.Errorf("token issue failed: %w", err)
	}

	return token, nil
}
