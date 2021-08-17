package api

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"

	"github.com/vbogretsov/guard/auth"
)

var (
	ErrUnexpectedProvider = echo.NewHTTPError(http.StatusBadRequest, "unexpected provider")
	ErrMissingCode        = echo.NewHTTPError(http.StatusBadRequest, "missing code")
)

type HealthCheck = func() error

type Factory interface {
	auth.Factory
	NewHealthCheck() HealthCheck
}

func New(h *HttpAPI) *echo.Echo {
	e := echo.New()
	e.GET("/:provider/callback", h.Callback)
	e.GET("/:provider", h.StartOAuth)
	e.POST("/refresh", h.Refresh)
	e.GET("/health", h.Health)

	e.HTTPErrorHandler = ErrorHandler
	return e
}

func ErrorHandler(err error, c echo.Context) {
	if errors.As(err, &auth.Error{}) {
		err = &echo.HTTPError{Code: http.StatusUnauthorized, Message: err}
	}
	c.Echo().DefaultHTTPErrorHandler(err, c)
}

type HttpAPI struct {
	factory Factory
}

func NewHttpAPI(factory Factory) *HttpAPI {
	return &HttpAPI{factory: factory}
}

func (h *HttpAPI) Callback(c echo.Context) error {
	provider, err := goth.GetProvider(c.Param("provider"))
	if err != nil {
		return ErrUnexpectedProvider
	}

	state := c.QueryParam("state")
	if state == "" {
		return ErrMissingCode
	}

	params := c.Request().URL.Query()

	token, err := h.factory.NewSignIner(provider).SignIn(state, params)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, token)
}

func (h *HttpAPI) Refresh(c echo.Context) error {
	token := c.FormValue("refresh_token")

	value, err := h.factory.NewRefresher().Refresh(token)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, value)
}

func (h *HttpAPI) StartOAuth(c echo.Context) error {
	provider, err := goth.GetProvider(c.Param("provider"))
	if err != nil {
		return ErrUnexpectedProvider
	}

	url, err := h.factory.NewOAuthStarter(provider).StartOAuth()
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *HttpAPI) Health(c echo.Context) error {
	hc := h.factory.NewHealthCheck()
	if err := hc(); err != nil {
		return err
	}
	return c.NoContent(http.StatusOK)
}
