package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func routeLoginPage() (string, string, echo.HandlerFunc) {
	return "GET", "/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	}
}

func routeRegisterPage() (string, string, echo.HandlerFunc) {
	return "GET", "/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register.html", nil)
	}
}
