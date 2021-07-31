package auth

import (
	"errors"
	"time"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

const xsrfLen = 128

type XSRFGenerator interface {
	Generate() (string, error)
}

type xsrfGenerator struct {
	tokens repo.XSRFTokens
	timer  Timer
	ttl    time.Duration
}

func NewXSRFGenerator(tokens repo.XSRFTokens, timer Timer, ttl time.Duration) XSRFGenerator {
	return &xsrfGenerator{
		tokens: tokens,
		timer:  timer,
		ttl:    ttl,
	}
}

func (c *xsrfGenerator) Generate() (string, error) {
	token, err := generateRandomString(refreshTokenLen)
	if err != nil {
		return "", err
	}

	now := c.timer.Now()

	record := model.XSRFToken{
		ID:      token,
		Created: now.Unix(),
		Expires: now.Add(c.ttl).Unix(),
	}

	if err := c.tokens.Create(record); err != nil {
		return "", err
	}

	return token, err
}

type XSRGValidator interface {
	Validate(token string) error
}

type xsrfValidator struct {
	tokens repo.XSRFTokens
	timer  Timer
}

func NewXSRFValidator(tokens repo.XSRFTokens, timer Timer) XSRGValidator {
	return &xsrfValidator{
		tokens: tokens,
		timer:  timer,
	}
}

func (c *xsrfValidator) Validate(token string) error {
	record, err := c.tokens.Find(token)
	if err != nil {
		if errors.Is(err, repo.ErrorNotFound) {
			return Error{msg: "invalid token"}
		}
		return err
	}

	if err := c.tokens.Delete(token); err != nil {
		return err
	}

	if record.Expires < c.timer.Now().Unix() {
		return Error{msg: "token expired"}
	}

	return nil
}
