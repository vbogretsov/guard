package main

import (
	"sort"
	"strings"
	"testing"

	"github.com/markbates/goth"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
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
			"P2_OIDC_CLIENT_ID=p2-id",
			"P2_OIDC_CLIENT_SECRET=p2-secret",
			"P3_OIDC_CLIENT_ID=p3-id",
			"OIDC_CLIENT_ID",
		}

		ps := addProviders(nil, environ)
		require.Equal(t, len(ps), 3)

		sort.Slice(ps, func(i, j int) bool {
			return strings.Compare(ps[i].name, ps[j].name) < 1
		})

		require.Equal(t, ps[0].name, "P1")
		require.Equal(t, ps[0].clientID(Conf{}), "p1-id")
		require.Equal(t, ps[0].clientSecret(Conf{}), "p1-secret")
		require.Equal(t, ps[1].name, "P2")
		require.Equal(t, ps[1].clientID(Conf{}), "p2-id")
		require.Equal(t, ps[1].clientSecret(Conf{}), "p2-secret")
		require.Equal(t, ps[2].name, "P3")
		require.Equal(t, ps[2].clientID(Conf{}), "p3-id")
		require.Equal(t, ps[2].clientSecret(Conf{}), "")
	})

}

func TestUseProviders(t *testing.T) {
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

	useProviders(cfg)

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
}
