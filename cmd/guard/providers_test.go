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
	cfg := Conf{
		BaseURL:          "http://localhost:8000",
		GoogleClientID:   "google-id",
		GoogleSecret:     "google-secret",
		FacebookClientID: "facebook-id",
	}

	zerolog.SetGlobalLevel(zerolog.Disabled)

	useProviders(cfg)

	g, err := goth.GetProvider("google")
	require.NoError(t, err)
	require.NotNil(t, g)
	require.Equal(t, g.Name(), "google")

	_, err = goth.GetProvider("facebook")
	require.Error(t, err)
}
