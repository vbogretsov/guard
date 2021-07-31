package auth

import (
	"github.com/markbates/goth"
)

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
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	xsrf, err := c.xsrf.Generate()
	if err != nil {
		return "", err
	}

	sess, err := provider.BeginAuth(xsrf)
	if err != nil {
		return "", err
	}

	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	return url, nil
}
