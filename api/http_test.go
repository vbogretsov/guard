package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/vbogretsov/guard/api"
	"github.com/vbogretsov/guard/auth"
)

type factoryMock struct {
	mock.Mock
}

func (m *factoryMock) NewSignIner(provider goth.Provider) auth.SignIner {
	return m.Called(provider).Get(0).(auth.SignIner)
}

func (m *factoryMock) NewRefresher() auth.Refresher {
	return m.Called().Get(0).(auth.Refresher)
}

func (m *factoryMock) NewOAuthStarter(provider goth.Provider) auth.OAuthStarter {
	return m.Called(provider).Get(0).(auth.OAuthStarter)
}

type signinerMock struct {
	mock.Mock
}

func (m *signinerMock) SignIn(code string, params goth.Params) (auth.Token, error) {
	args := m.Called(code, params)

	v := args.Get(0)
	if v == nil {
		return auth.Token{}, args.Error(1)
	}

	return v.(auth.Token), args.Error(1)
}

type refresherMock struct {
	mock.Mock
}

func (m *refresherMock) Refresh(token string) (auth.Token, error) {
	args := m.Called(token)

	v := args.Get(0)
	if v == nil {
		return auth.Token{}, args.Error(1)
	}

	return v.(auth.Token), args.Error(1)
}

type oauthStarterMock struct {
	mock.Mock
}

func (m *oauthStarterMock) StartOAuth() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

type context struct {
	e            *echo.Echo
	c            echo.Context
	factory      *factoryMock
	signiner     *signinerMock
	refresher    *refresherMock
	oauthStarter *oauthStarterMock
	handler      *api.HttpAPI
	req          *http.Request
	rec          *httptest.ResponseRecorder
}

func newctx(path string) *context {
	e := echo.New()

	factory := &factoryMock{}
	signiner := &signinerMock{}
	refresher := &refresherMock{}
	oauthStarter := &oauthStarterMock{}

	handler := api.NewHttpAPI(factory)

	factory.On("NewSignIner", mock.Anything).Return(signiner)
	factory.On("NewRefresher", mock.Anything).Return(refresher)
	factory.On("NewOAuthStarter", mock.Anything).Return(oauthStarter)

	api.Setup(e, handler)

	req := httptest.NewRequest(http.MethodGet, path, nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	c.SetPath(path)

	return &context{
		e:            e,
		c:            c,
		factory:      factory,
		signiner:     signiner,
		refresher:    refresher,
		oauthStarter: oauthStarter,
		handler:      handler,
		req:          req,
		rec:          rec,
	}
}

func init() {
	goth.UseProviders(google.New("google_id", "google_secret", "http://localhost:8000/google/callback"))
}

func TestErrorhandler(t *testing.T) {
	t.Run("401", func(t *testing.T) {
		ctx := newctx("/")
		api.ErrorHandler(auth.Error{}, ctx.c)
		require.Equal(t, http.StatusUnauthorized, ctx.rec.Code)
	})
	t.Run("500", func(t *testing.T) {
		ctx := newctx("/")
		api.ErrorHandler(errors.New("unexpected error"), ctx.c)
		require.Equal(t, http.StatusInternalServerError, ctx.rec.Code)
	})
	t.Run("ErrUnexpectedProvider", func(t *testing.T) {
		ctx := newctx("/")
		api.ErrorHandler(api.ErrUnexpectedProvider, ctx.c)
		require.Equal(t, http.StatusBadRequest, ctx.rec.Code)
	})
	t.Run("ErrMissingCode", func(t *testing.T) {
		ctx := newctx("/")
		api.ErrorHandler(api.ErrMissingCode, ctx.c)
		require.Equal(t, http.StatusBadRequest, ctx.rec.Code)
	})
}

func TestHttpStartOAuth(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := newctx("/:provider")
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("google")

		redirectURL := "redirectURL"
		ctx.oauthStarter.On("StartOAuth").Return(redirectURL, nil)

		err := ctx.handler.StartOAuth(ctx.c)
		require.NoError(t, err)
		require.Equal(t, http.StatusTemporaryRedirect, ctx.rec.Code)
		require.Equal(t, redirectURL, ctx.rec.HeaderMap["Location"][0])
	})

	t.Run("BadProvider", func(t *testing.T) {
		ctx := newctx("/:provider")
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("xxx")

		err := ctx.handler.StartOAuth(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, api.ErrUnexpectedProvider)
	})

	t.Run("InternalError", func(t *testing.T) {
		ctx := newctx("/:provider")
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("google")

		fail := errors.New("unexpected error")
		ctx.oauthStarter.On("StartOAuth").Return("", fail)

		err := ctx.handler.StartOAuth(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}

func TestHttpSignIn(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		code := "signin123"
		q := make(url.Values)
		q.Set("code", code)

		ctx := newctx("/:provider/callback/?" + q.Encode())
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("google")

		token := auth.Token{
			IssuedAt:       1600000000,
			Access:         "access.123",
			AccessExpires:  1600000050,
			Refresh:        "refresh.123",
			RefreshExpires: 1600000100,
		}

		ctx.signiner.On("SignIn", mock.Anything, mock.Anything).Return(token, nil)

		err := ctx.handler.Callback(ctx.c)
		require.NoError(t, err)

		var value auth.Token
		require.NoError(t, json.Unmarshal(ctx.rec.Body.Bytes(), &value))
		require.Equal(t, token, value)
	})

	t.Run("InvalidProvider", func(t *testing.T) {
		ctx := newctx("/:provider/callback")
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("xxx")

		err := ctx.handler.Callback(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, api.ErrUnexpectedProvider)
	})

	t.Run("MissingCode", func(t *testing.T) {
		ctx := newctx("/:provider/callback")
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("google")

		err := ctx.handler.Callback(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, api.ErrMissingCode)
	})

	t.Run("InternalError", func(t *testing.T) {
		code := "signin123"
		q := make(url.Values)
		q.Set("code", code)

		ctx := newctx("/:provider/callback/?" + q.Encode())
		ctx.c.SetParamNames("provider")
		ctx.c.SetParamValues("google")

		fail := errors.New("unexpected error")
		ctx.signiner.On("SignIn", mock.Anything, mock.Anything).Return(nil, fail)

		err := ctx.handler.Callback(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}

func TestHttpRefresh(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		refreshToken := "refresh.123"

		form := make(url.Values)
		form.Set("refresh_token", refreshToken)

		ctx := newctx("/refresh")

		token := auth.Token{
			IssuedAt:       1600000000,
			Access:         "access.123",
			AccessExpires:  1600000050,
			Refresh:        "refresh.456",
			RefreshExpires: 1600000100,
		}

		ctx.refresher.On("Refresh", refreshToken).Return(token, nil)
		ctx.req.Form = form

		err := ctx.handler.Refresh(ctx.c)
		require.NoError(t, err)

		var value auth.Token
		require.NoError(t, json.Unmarshal(ctx.rec.Body.Bytes(), &value))
		require.Equal(t, token, value)
	})

	t.Run("Expired", func(t *testing.T) {
		refreshToken := "refresh.123"

		form := make(url.Values)
		form.Set("refresh_token", refreshToken)

		ctx := newctx("/refresh")

		fail := auth.Error{}
		ctx.refresher.On("Refresh", refreshToken).Return(nil, fail)
		ctx.req.Form = form

		err := ctx.handler.Refresh(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})

	t.Run("Invalid", func(t *testing.T) {
		ctx := newctx("/refresh")

		fail := auth.Error{}
		ctx.refresher.On("Refresh", "").Return(nil, fail)

		err := ctx.handler.Refresh(ctx.c)
		require.Error(t, err)
		require.ErrorIs(t, err, fail)
	})
}
