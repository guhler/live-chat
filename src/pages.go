package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

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
