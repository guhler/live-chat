package routes

import (
	"database/sql"
	"live_chat/util"
	"net/http"

	"github.com/labstack/echo/v4"
)

func GetIndex(db *sql.DB) (string, string, echo.HandlerFunc) {
	return "GET", "/", func(c echo.Context) error {
		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.Render(http.StatusOK, "index.html", nil)
		}

		roomNames, err := util.GetRoomsOfUser(db, userName)
		if err != nil {
			return err
		}

		return c.Render(http.StatusOK, "index_auth.html", indexPage{roomNames})
	}
}
