package auth

import "github.com/vbogretsov/guard/repo"

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
		tx:     tx,
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
