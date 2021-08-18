package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/markbates/goth"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"gopkg.in/h2non/gock.v1"
)

func TestAddProviders(t *testing.T) {
	t.Run("Empty", func(t *testing.T) {
		ps := addProviders(nil, []string{})
		require.Equal(t, len(ps), 0)
	})

	t.Run("NotEmpty", func(t *testing.T) {
		environ := []string{
			"P1_OIDC_CLIENT_ID=p1-id",
			"P1_OIDC_CLIENT_SECRET=p1-secret",
			"P1_OIDC_DISCOVERY_URL=http://p1.org/discovery",
			"P2_OIDC_CLIENT_ID=p2-id",
			"P2_OIDC_CLIENT_SECRET=p2-secret",
			"P2_OIDC_DISCOVERY_URL=http://p2.org/discovery",
			"P3_OIDC_CLIENT_ID=p3-id",
			"OIDC_CLIENT_ID",
		}

		ps := addProviders(nil, environ)
		require.Equal(t, len(ps), 2)

		sort.Slice(ps, func(i, j int) bool {
			return strings.Compare(ps[i].name, ps[j].name) < 1
		})

		require.Equal(t, ps[0].name, "P1")
		require.Equal(t, ps[0].clientID(Conf{}), "p1-id")
		require.Equal(t, ps[0].clientSecret(Conf{}), "p1-secret")
		require.Equal(t, ps[1].name, "P2")
		require.Equal(t, ps[1].clientID(Conf{}), "p2-id")
		require.Equal(t, ps[1].clientSecret(Conf{}), "p2-secret")
	})

}

func TestUseProviders(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cfg := Conf{
			BaseURL:           "http://localhost:8000",
			AppleClientID:     "apple-id",
			AppleClientSecret: "apple-secret",
			GoogleClientID:    "google-id",
			GoogleSecret:      "google-secret",
			FacebookClientID:  "facebook-id",
			FacebookSecret:    "faacebook-secret",
			TwitterClientID:   "twitter-id",
			TwitterSecret:     "twitter-secret",
			VkClientID:        "vk-id",
			VkSecret:          "vk-secret",
			YandexClientID:    "yandex-id",
			YandexSecret:      "yandex-secret",
		}

		zerolog.SetGlobalLevel(zerolog.Disabled)

		useProviders(cfg, []string{})

		a, err := goth.GetProvider("apple")
		require.NoError(t, err)
		require.NotNil(t, a)

		g, err := goth.GetProvider("google")
		require.NoError(t, err)
		require.NotNil(t, g)

		f, err := goth.GetProvider("facebook")
		require.NoError(t, err)
		require.NotNil(t, f)

		tw, err := goth.GetProvider("twitter")
		require.NoError(t, err)
		require.NotNil(t, tw)

		v, err := goth.GetProvider("vk")
		require.NoError(t, err)
		require.NotNil(t, v)

		y, err := goth.GetProvider("yandex")
		require.NoError(t, err)
		require.NotNil(t, y)

		goth.ClearProviders()
	})

	t.Run("Missconfigured", func(t *testing.T) {
		cfg := Conf{
			BaseURL:        "http://localhost:8000",
			VkClientID:     "",
			VkSecret:       "vk-secret",
			YandexClientID: "yandex-id",
			YandexSecret:   "",
		}

		zerolog.SetGlobalLevel(zerolog.Disabled)

		useProviders(cfg, []string{})

		var err error

		_, err = goth.GetProvider("vk")
		require.Error(t, err)

		_, err = goth.GetProvider("yandex")
		require.Error(t, err)

		goth.ClearProviders()
	})

	t.Run("OpenIdConnectSuccess", func(t *testing.T) {
		defer gock.Off()

		cfg := Conf{
			BaseURL: "http://localhost:8000",
		}

		zerolog.SetGlobalLevel(zerolog.Disabled)

		discoveryURL := "http://p1oidc.org"
		gock.New(discoveryURL).
			Get("/discovery").
			Reply(200).
			JSON(map[string]string{
				"authorization_endpoint": "/authorize",
				"token_endpoint":         "/token",
				"userinfo_endpoint":      "/userinfo",
				"issuer":                 "p1",
			})

		useProviders(cfg, []string{
			"P1_OIDC_CLIENT_ID=p1-id",
			"P1_OIDC_CLIENT_SECRET=p1-secret",
			fmt.Sprintf("P1_OIDC_DISCOVERY_URL=%s/discovery", discoveryURL),
		})

		p1, err := goth.GetProvider("openid-connect")
		require.NoError(t, err)
		require.NotNil(t, p1)

		goth.ClearProviders()
	})

	t.Run("OpenIdConnectFailed", func(t *testing.T) {
		defer gock.Off()

		cfg := Conf{
			BaseURL: "http://localhost:8000",
		}

		zerolog.SetGlobalLevel(zerolog.Disabled)

		discoveryURL := "http://p1oidc.org"
		gock.New(discoveryURL).
			Get("/discovery").
			Reply(404)

		useProviders(cfg, []string{
			"P1_OIDC_CLIENT_ID=p1-id",
			"P1_OIDC_CLIENT_SECRET=p1-secret",
			fmt.Sprintf("P1_OIDC_DISCOVERY_URL=%s/discovery", discoveryURL),
		})

		_, err := goth.GetProvider("openid-connect")
		require.Error(t, err)

		goth.ClearProviders()
	})
}
