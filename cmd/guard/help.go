package main

import (
	"flag"
	"fmt"
	"os"
)

var Version = "latest"

var header = fmt.Sprintf("guard -- authentication microservice (%s)\n", Version)

const usage = `
Configuration environment variables:

	GUARD_PORT
		TCP port to listen. Default: 8000
	GUARD_DSN
		Database connection string. Required.
		Example: postgres://username:password@host:port/database
	GUARD_SECRET_KEY
		Secret key used for tokens encryption.
	GUARD_ACCESS_TTL
		Access token TTL. Default 300s
	GUARD_REFRESH_TTL
		Refresh token TTL. Default 86400s
	GUARD_CALLBACK_URL
		Callback URL used during OAuth authentication process.
		Default: http://localhost:8000/callback

Wellknown OAuth providers environment variables:

	APPLE_CLIENT_ID, APPLE_CLIENT_SECRET       -- Apple
	GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET     -- Google
	FACEBOOK_CLIENT_ID, FACEBOOK_CLIENT_SECRET -- Facebook
	TWITTER_CLIENT_ID, TWITTER_CLIENT_SECRET   -- Twitter
	VK_CLIENT_ID, VK_CLIENT_SECRET             -- Vk
	YANDEX_CLIENT_ID, YANDEX_CLIENT_SECRET     -- Yandex

Custom OIDC providers variables can also be passed:

	PROVIDER_1_OIDC_CLIENT_ID, PROVIDER_1_OIDC_CLIENT_SECRET
	...
	PROVIDER_N_OIDC_CLIENT_ID, PROVIDER_N_OIDC_CLIENT_SECRET

where PROVIDER_1, ..., PROVIDER_N -- just any prefixes used to group client id
and client secret for a particular provider.
`

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Println(header + usage)
		flag.PrintDefaults()
	}
}
