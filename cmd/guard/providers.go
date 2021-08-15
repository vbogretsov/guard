package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/apple"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/openidConnect"
	"github.com/markbates/goth/providers/twitter"
	"github.com/markbates/goth/providers/vk"
	"github.com/markbates/goth/providers/yandex"
	"github.com/rs/zerolog/log"
)

const (
	oidcIdSuffix = "_OIDC_CLIENT_ID"
	/* #nosec G101 */
	oidcSecretSuffix = "_OIDC_CLIENT_SECRET"
)

type provider struct {
	name         string
	ctor         func(string, string, string) goth.Provider
	clientID     func(cfg Conf) string
	clientSecret func(cfg Conf) string
}

var providers = []provider{
	{
		name:         "apple",
		ctor:         func(id, secret, url string) goth.Provider { return apple.New(id, secret, url, nil) },
		clientID:     func(cfg Conf) string { return cfg.AppleClientID },
		clientSecret: func(cfg Conf) string { return cfg.AppleClientSecret },
	},
	{
		name:         "google",
		ctor:         func(id, secret, url string) goth.Provider { return google.New(id, secret, url) },
		clientID:     func(cfg Conf) string { return cfg.GoogleClientID },
		clientSecret: func(cfg Conf) string { return cfg.GoogleSecret },
	},
	{
		name:         "facebook",
		ctor:         func(id, secret, url string) goth.Provider { return facebook.New(id, secret, url) },
		clientID:     func(cfg Conf) string { return cfg.FacebookClientID },
		clientSecret: func(cfg Conf) string { return cfg.FacebookSecret },
	},
	{
		name:         "twitter",
		ctor:         func(id, secret, url string) goth.Provider { return twitter.New(id, secret, url) },
		clientID:     func(cfg Conf) string { return cfg.TwitterClientID },
		clientSecret: func(cfg Conf) string { return cfg.TwitterSecret },
	},
	{
		name:         "vk",
		ctor:         func(id, secret, url string) goth.Provider { return vk.New(id, secret, url) },
		clientID:     func(cfg Conf) string { return cfg.VkClientID },
		clientSecret: func(cfg Conf) string { return cfg.VkSecret },
	},
	{
		name:         "yandex",
		ctor:         func(id, secret, url string) goth.Provider { return yandex.New(id, secret, url) },
		clientID:     func(cfg Conf) string { return cfg.YandexClientID },
		clientSecret: func(cfg Conf) string { return cfg.YandexSecret },
	},
}

func newOpenIDProvider(id, secret, url string) goth.Provider {
	provider, err := openidConnect.New(id, secret, url, "")
	if err != nil {
		panic(fmt.Errorf("unable to create provider: %w", err))
	}
	return provider
}

func addProviders(providers []provider, environ []string) []provider {
	envs := map[string]string{}
	for _, e := range environ {
		kv := strings.Split(e, "=")
		if len(kv) != 2 {
			continue
		}
		envs[kv[0]] = kv[1]
	}

	for k := range envs {
		ind := strings.Index(k, oidcIdSuffix)
		if ind == -1 {
			continue
		}

		name := k[:ind]

		providers = append(providers, provider{
			name:         k[:ind],
			ctor:         newOpenIDProvider,
			clientID:     func(cfg Conf) string { return envs[name+oidcIdSuffix] },
			clientSecret: func(cfg Conf) string { return envs[name+oidcSecretSuffix] },
		})
	}

	return providers
}

func useProviders(cfg Conf) {
	for _, p := range addProviders(providers, os.Environ()) {
		clientID := p.clientID(cfg)
		if clientID == "" {
			log.Warn().
				Str("provider", p.name).
				Str("reason", "missing client id").
				Msg("unable to use provider")
			continue
		}
		clientSecret := p.clientSecret(cfg)
		if clientSecret == "" {
			log.Warn().
				Str("provider", p.name).
				Str("reason", "missing client secret").
				Msg("unable to use provider")
			continue
		}
		callbackURL := fmt.Sprintf("%s/%s/callback", cfg.BaseURL, p.name)
		goth.UseProviders(p.ctor(clientID, clientSecret, callbackURL))
	}
}
