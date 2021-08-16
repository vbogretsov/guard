package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDbConn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		db, err := dbconnect(Conf{
			DBDriver:           "sqlite",
			DSN:                ":memory:",
			DBMaxOpenConn:      10,
			DBMaxIddleConn:     10,
			DBConnMaxLifetime:  1 * time.Hour,
			DBConnMaxIddleTime: 1 * time.Hour,
		})

		require.NoError(t, err)
		require.NotNil(t, db)
	})
	t.Run("UnsupportedProvider", func(t *testing.T) {
		_, err := dbconnect(Conf{
			DBDriver: "xxx",
			DSN:      ":memory:",
		})

		require.Error(t, err)
	})
	t.Run("InvalidConnection", func(t *testing.T) {
		_, err := dbconnect(Conf{
			DBDriver: "postgres",
			DSN:      "postgres://user:password@localhost/test",
		})

		require.Error(t, err)
	})
}
