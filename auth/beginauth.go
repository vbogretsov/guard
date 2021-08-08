package auth

import (
	"time"

	"github.com/markbates/goth"
	"github.com/vbogretsov/guard/model"
	"github.com/vbogretsov/guard/repo"
)

const codeLen = 128

type OAuthStarter interface {
	StartOAuth() (string, error)
}

type oauthStarter struct {
	ttl      time.Duration
	timer    Timer
	sessions repo.Sessions
	provider goth.Provider
}

func NewOAuthStarter(ttl time.Duration, timer Timer, sessions repo.Sessions, provider goth.Provider) OAuthStarter {
	return &oauthStarter{
		ttl:      ttl,
		timer:    timer,
		sessions: sessions,
		provider: provider,
	}
}

func (c *oauthStarter) StartOAuth() (string, error) {
	code, err := generateRandomString(codeLen)
	if err != nil {
		return "", err
	}

	sess, err := c.provider.BeginAuth(code)
	if err != nil {
		return "", err
	}

	now := c.timer.Now()

	record := model.Session{
		ID:      code,
		Value:   sess.Marshal(),
		Created: now.Unix(),
		Expires: now.Add(c.ttl).Unix(),
	}

	if err := c.sessions.Create(record); err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	return url, nil
}
