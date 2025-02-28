package routes

import (
	"database/sql"
	"live_chat/auth"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetIndex(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "GET", "/", func(c echo.Context) error {
		_, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.Render(http.StatusOK, "index.html", nil)
		}

		return c.Redirect(http.StatusFound, "/rooms")
	}, auth.TokenParser
}
