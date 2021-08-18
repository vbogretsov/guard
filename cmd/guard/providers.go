package main

import (
	"fmt"
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
	oidcURLSuffix = "_OIDC_DISCOVERY_URL"
	oidcIdSuffix  = "_OIDC_CLIENT_ID"
	/* #nosec G101 */
	oidcSecretSuffix = "_OIDC_CLIENT_SECRET"
)

type provider struct {
	name         string
	ctor         func(string, string, string) (goth.Provider, error)
	clientID     func(cfg Conf) string
	clientSecret func(cfg Conf) string
}

var providers = []provider{
	{
		name:         "apple",
		ctor:         func(id, secret, url string) (goth.Provider, error) { return apple.New(id, secret, url, nil), nil },
		clientID:     func(cfg Conf) string { return cfg.AppleClientID },
		clientSecret: func(cfg Conf) string { return cfg.AppleClientSecret },
	},
	{
		name:         "google",
		ctor:         func(id, secret, url string) (goth.Provider, error) { return google.New(id, secret, url), nil },
		clientID:     func(cfg Conf) string { return cfg.GoogleClientID },
		clientSecret: func(cfg Conf) string { return cfg.GoogleSecret },
	},
	{
		name:         "facebook",
		ctor:         func(id, secret, url string) (goth.Provider, error) { return facebook.New(id, secret, url), nil },
		clientID:     func(cfg Conf) string { return cfg.FacebookClientID },
		clientSecret: func(cfg Conf) string { return cfg.FacebookSecret },
	},
	{
		name:         "twitter",
		ctor:         func(id, secret, url string) (goth.Provider, error) { return twitter.New(id, secret, url), nil },
		clientID:     func(cfg Conf) string { return cfg.TwitterClientID },
		clientSecret: func(cfg Conf) string { return cfg.TwitterSecret },
	},
	{
		name:         "vk",
		ctor:         func(id, secret, url string) (goth.Provider, error) { return vk.New(id, secret, url), nil },
		clientID:     func(cfg Conf) string { return cfg.VkClientID },
		clientSecret: func(cfg Conf) string { return cfg.VkSecret },
	},
	{
		name:         "yandex",
		ctor:         func(id, secret, url string) (goth.Provider, error) { return yandex.New(id, secret, url), nil },
		clientID:     func(cfg Conf) string { return cfg.YandexClientID },
		clientSecret: func(cfg Conf) string { return cfg.YandexSecret },
	},
}

func oidcProvider(discoveryURL string) func(id, secret, url string) (goth.Provider, error) {
	return func(id, secret, url string) (goth.Provider, error) {
		return openidConnect.New(id, secret, url, discoveryURL)
	}
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

		discoveryURL, ok := envs[name+oidcURLSuffix]
		if !ok {
			log.Warn().
				Str("name", name).
				Str("cause", "missing discovery URL").
				Msg("failed to register OIDC provider")
			continue
		}

		providers = append(providers, provider{
			name:         k[:ind],
			ctor:         oidcProvider(discoveryURL),
			clientID:     func(cfg Conf) string { return envs[name+oidcIdSuffix] },
			clientSecret: func(cfg Conf) string { return envs[name+oidcSecretSuffix] },
		})
	}

	return providers
}

func useProviders(cfg Conf, environ []string) {
	for _, p := range addProviders(providers, environ) {
		clientID := p.clientID(cfg)
		if clientID == "" {
			log.Warn().
				Str("provider", p.name).
				Str("cause", "missing client id").
				Msg("failed to use provider")
			continue
		}

		clientSecret := p.clientSecret(cfg)
		if clientSecret == "" {
			log.Warn().
				Str("provider", p.name).
				Str("cause", "missing client secret").
				Msg("failed to use provider")
			continue
		}

		callbackURL := fmt.Sprintf("%s/%s/callback", cfg.BaseURL, p.name)

		pvr, err := p.ctor(clientID, clientSecret, callbackURL)
		if err != nil {
			log.Warn().
				Str("provider", p.name).
				Str("cause", err.Error()).
				Msg("failed to use provider")
			continue
		}

		goth.UseProviders(pvr)
	}
}
