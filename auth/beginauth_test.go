package auth_test

import (
	"testing"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/stretchr/testify/require"
	"github.com/vbogretsov/guard/auth"
)

func init() {
	goth.UseProviders(
		google.New("GOOGLE_KEY", "GOOGLE_SECRET", "http://localhost:8000/auth/google/callback"),
	)
}

func TestStartOAuth(t *testing.T) {
	xsrf := &xsrfGeneratorMock{}

	xsrf.On("Generate").Return("beginauth.xsrf.123", nil)

	cmd := auth.NewOAuthStarter(xsrf)

	result, err := cmd.StartOAuth("google")
	require.NoError(t, err)
	require.NotEmpty(t, result)
}
