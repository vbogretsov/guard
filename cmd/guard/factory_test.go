package main

import (
	"testing"

	"github.com/markbates/goth/providers/google"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFactory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	pr := google.New("google_id", "google_secret", "http://localhost:8000/google/callback")

	factory := NewFactory(db, FactoryConfig{})
	require.NotNil(t, factory.NewSignIner(pr))
	require.NotSame(t, factory.NewSignIner(pr), factory.NewSignIner(pr))
	require.NotNil(t, factory.NewOAuthStarter(pr))
	require.NotSame(t, factory.NewOAuthStarter(pr), factory.NewOAuthStarter(pr))
	require.NotNil(t, factory.NewRefresher())
	require.NotSame(t, factory.NewRefresher(), factory.NewRefresher())

	require.NoError(t, factory.NewHealthCheck()())
}
