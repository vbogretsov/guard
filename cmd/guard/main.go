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
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/ziflex/lecho"

	"github.com/vbogretsov/guard/api"
)

func run(sig chan os.Signal) error {
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

	useProviders(&cfg)

	e := echo.New()
	e.Debug = cfg.Debug
	e.HideBanner = true
	e.Logger = lecho.New(os.Stdout)

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.Logger())

	e.GET("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	httpAPI := api.NewHttpAPI(NewFactory(db, FactoryConfig{
		SecretKey:  cfg.SecretKey,
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
		CodeTTL:    cfg.CodeTTL,
	}))

	api.Setup(e, httpAPI)

	exit := make(chan error)
	go func() {
		exit <- e.Start(fmt.Sprintf(":%d", cfg.Port))
	}()

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

	sig := make(chan os.Signal, 1)
	if err := run(sig); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
