package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

type Server interface {
	Start(string) error
	Shutdown(context.Context) error
}

func start(server Server, address string, sig chan os.Signal, timeout time.Duration) error {
	exit := make(chan error)
	go func() {
		exit <- server.Start(address)
	}()

	signal.Notify(sig, syscall.SIGTERM)
	<-sig

	log.Info().Msg("received SIGTERM")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Info().Msg("terminating")
	if err := server.Shutdown(ctx); err != nil {
		return err
	}

	return <-exit
}
