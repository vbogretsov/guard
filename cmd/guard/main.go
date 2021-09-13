package main

import (
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/caarlos0/env"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/ziflex/lecho"

	"github.com/vbogretsov/guard/api"
)

const shutdownTimeout = 10 * time.Second

func run() error {
	cfg := Conf{}
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	db, err := dbconnect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	logLevel, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("unable to parse log level: %w", err)
	}

	zerolog.SetGlobalLevel(logLevel)

	useProviders(cfg, os.Environ())

	h := api.NewHttpAPI(NewFactory(db, FactoryConfig{
		SecretKey:  cfg.SecretKey,
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
		CodeTTL:    cfg.CodeTTL,
	}))

	e := api.New(h)
	e.Debug = cfg.Debug
	e.HideBanner = true
	e.Logger = lecho.New(os.Stdout)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())

	sig := make(chan os.Signal, 1)
	return start(e, fmt.Sprintf(":%d", cfg.Port), sig, shutdownTimeout)
}

func main() {
	flag.Parse()

	if err := run(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
