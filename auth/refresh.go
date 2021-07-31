package auth

import (
	"errors"
	"time"

	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

const refreshTokenLen = 128

type RefreshTokenCreator interface {
	Create(user model.User) (model.RefreshToken, error)
}

type refreshTokenCreator struct {
	tokens repo.RefreshTokens
	timer  Timer
	ttl    time.Duration
}

func NewRefreshTokenCreator(tokens repo.RefreshTokens, timer Timer, ttl time.Duration) RefreshTokenCreator {
	return &refreshTokenCreator{
		tokens: tokens,
		timer:  timer,
		ttl:    ttl,
	}
}

func (c *refreshTokenCreator) Create(user model.User) (model.RefreshToken, error) {
	refreshID, err := generateRandomString(refreshTokenLen)
	if err != nil {
		return model.RefreshToken{}, err
	}

	now := c.timer.Now()

	token := model.RefreshToken{
		ID:      refreshID,
		UserID:  user.ID,
		User:    user,
		Created: now.Unix(),
		Expires: now.Add(c.ttl).Unix(),
	}

	if err := c.tokens.Create(token); err != nil {
		return token, err
	}

	return token, nil
}

type Refresher interface {
	Refresh(refreshToken string) (Token, error)
}

type refresher struct {
	tx     repo.Transaction
	timer  Timer
	tokens repo.RefreshTokens
	issuer Issuer
}

func NewRefresher(timer Timer, tx repo.Transaction, tokens repo.RefreshTokens, issuer Issuer) Refresher {
	return &refresher{
		tx:     tx,
		timer:  timer,
		tokens: tokens,
		issuer: issuer,
	}
}

func (c *refresher) Refresh(refreshToken string) (Token, error) {
	var empty Token

	if err := c.tx.Begin(); err != nil {
		return empty, err
	}
	defer c.tx.Close()

	old, err := c.tokens.Find(refreshToken)
	if err != nil {
		if errors.Is(err, repo.ErrorNotFound) {
			return empty, Error{msg: "invalid token"}
		}
		return empty, err
	}

	if err := c.tokens.Delete(old.ID); err != nil {
		return empty, err
	}

	if old.Expires < c.timer.Now().Unix() {
		return empty, Error{msg: "expired token"}
	}

	token, err := c.issuer.Issue(old.User)
	if err != nil {
		return empty, err
	}

	if err := c.tx.Commit(); err != nil {
		return empty, err
	}

	return token, nil
}
