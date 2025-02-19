package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func routeIndex(e *echo.Echo) error {
	e.GET("/", func(c echo.Context) error {

		if user, ok := c.Get("authorized_user").(string); ok && user != "" {
			return c.Render(http.StatusOK, "index_auth.html", nil)
		} else {
			return c.Render(http.StatusOK, "index.html", nil)
		}
	})
	return nil
}

func routeLoginPage(e *echo.Echo) error {
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	})
	return nil
}

func routeRegisterPage(e *echo.Echo) error {
	e.GET("/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register.html", nil)
	})
	return nil
}
