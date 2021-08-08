package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/markbates/goth"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/markbates/goth/providers/google"

	"github.com/vbogretsov/guard/api"
	"github.com/vbogretsov/guard/auth"
)

func setup(cfg Conf) {
	if cfg.GoogleClientID == "" {
		log.Warn().
			Str("reason", "missing client id").
			Str("env", "GOOGLE_CLIENT_ID").
			Msg("failed to configure google provider")
		return
	}

	if cfg.GoogleClientSecret == "" {
		log.Warn().
			Str("reason", "missing client id").
			Str("env", "GOOGLE_CLIENT_SECRET").
			Msg("failed to configure google provider")
		return
	}

	goth.UseProviders(
		google.New(cfg.GoogleClientID, cfg.GoogleClientSecret, fmt.Sprintf("%s/google/callback", cfg.BaseURL)),
	)

	log.Info().Msg("initialized google provider")
}

func run() error {
	cfg := Conf{}
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	setup(cfg)

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	factory := auth.NewFactory(db, auth.Config{
		SecretKey:  cfg.SecretKey,
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
		CodeTTL:    cfg.CodeTTL,
	})

	httpAPI := api.NewHttpAPI(factory)

	e := echo.New()
	e.Pre(middleware.RemoveTrailingSlash())

	e.GET("/:provider/callback", httpAPI.Callback)
	e.GET("/:provider", httpAPI.StartOAuth)
	e.POST("/refresh", httpAPI.Refresh)

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	e.Debug = true
	e.HideBanner = true

	e.Logger = lecho.New(os.Stdout)
	e.Use(middleware.Logger())

	exit := make(chan error)
	go func() {
		exit <- e.Start(fmt.Sprintf(":%d", cfg.Port))
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM)
	<-sig

	log.Info().Msg("received SIGTERM")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info().Msg("terminating")
	if err := e.Shutdown(ctx); err != nil {
		return err
	}

	return <-exit
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
