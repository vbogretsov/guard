package main

import (
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

func TestIntegration(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.Disabled)

	sig := make(chan os.Signal)
	done := make(chan error, 1)
	go func(t *testing.T) {
		err := run(sig)
		done <- err
	}(t)

	wait := 10

	ok := false
	for i := 0; i < wait; i++ {
		resp, err := http.Get("http://localhost:8000/health")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		require.Equal(t, http.StatusOK, resp.StatusCode)
		ok = true
		break
	}

	require.Equal(t, ok, true, "server not started in %d seconds", wait)

	sig <- syscall.SIGTERM

	ticker := time.NewTicker(time.Duration(wait) * time.Second)
	select {
	case <-ticker.C:
		t.Errorf("not stopped after %v", ticker)
	case err := <-done:
		require.Error(t, err)
		require.ErrorIs(t, err, http.ErrServerClosed)
	}
}
