package auth

import (
	"errors"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

type SessionValidator interface {
	Validate(token string) (model.Session, error)
}

type sessionValidator struct {
	sessions repo.Sessions
	timer    Timer
}

func NewSessionValidator(sessions repo.Sessions, timer Timer) SessionValidator {
	return &sessionValidator{
		sessions: sessions,
		timer:    timer,
	}
}

func (c *sessionValidator) Validate(code string) (model.Session, error) {
	var value model.Session

	record, err := c.sessions.Find(code)
	if err != nil {
		if errors.Is(err, repo.ErrorNotFound) {
			return value, Error{msg: "invalid code"}
		}
		return value, err
	}

	if record.Expires < c.timer.Now().Unix() {
		return value, Error{msg: "session expired"}
	}

	return record, nil
}
