package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

type serverMock struct {
	running chan bool
	delay   time.Duration
}

func (m *serverMock) Start(address string) error {
	<-m.running
	return nil
}

func (m *serverMock) Shutdown(ctx context.Context) error {
	time.Sleep(m.delay)
	m.running <- false
	return ctx.Err()
}

func TestGracefullShutdown(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)
	timeout := 1 * time.Second

	t.Run("Success", func(t *testing.T) {
		srv := &serverMock{running: make(chan bool)}
		sig := make(chan os.Signal, 1)

		done := make(chan error, 1)
		go func() {
			done <- start(srv, "", sig, timeout)
		}()

		sig <- syscall.SIGTERM
		tic := time.NewTicker(timeout)

		select {
		case <-tic.C:
			t.Errorf("not stopped after %v", timeout)
		case err := <-done:
			require.NoError(t, err)
		}
	})

	t.Run("TimedOut", func(t *testing.T) {
		srv := &serverMock{running: make(chan bool), delay: timeout * 2}
		sig := make(chan os.Signal, 1)

		done := make(chan error, 1)
		go func() {
			done <- start(srv, "", sig, timeout)
		}()

		sig <- syscall.SIGTERM

		threshold := timeout * 3
		tic := time.NewTicker(timeout * 3)

		select {
		case <-tic.C:
			t.Errorf("not stopped after %v", threshold)
		case err := <-done:
			require.Error(t, err)
		}
	})
}
