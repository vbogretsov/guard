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

	LOGIND_PORT
		TCP port to listen. Default: 8000
	LOGIND_DSN
		Database connection string. Required.
		Example: postgres://username:password@host:port/database
	LOGIND_SECRET_KEY
		Secret key used for tokens encryption.
	LOGIND_ACCESS_TTL
		Access token TTL. Default 300s
	LOGIND_REFRESH_TTL
		Refresh token TTL. Default 86400s
	LOGIND_CALLBACK_URL
		Callback URL used during OAuth authentication process.
		Default: http://localhost:8000/callback

OAuth providers environment variables:

	GOOGLE_CLIENT_ID
		Google client ID
	GOOGLE_CLIENT_SECRET
		Google client secret
`

func init() {
	flag.CommandLine.SetOutput(os.Stdout)
	flag.Usage = func() {
		fmt.Println(header + usage)
		flag.PrintDefaults()
	}
}
