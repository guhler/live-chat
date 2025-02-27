package auth

import (
	"database/sql"
	"live_chat/util"
	"net/http"

	"github.com/labstack/echo/v4"
)

func TokenParser(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("auth-token")
		if err != nil {
			if err != echo.ErrCookieNotFound {
				c.Logger().Warn("Cookie Error: ", err)
			}
			return next(c)
		}

		tk := cookie.Value
		username, err := ValidateToken(tk)
		if err != nil {
			c.Logger().Warn("Failed to validate Token: ", err)
			return next(c)
		}
		if util.ValidateUserName(username) != nil {
			c.Logger().Warn("Parsed invalid user name from Token: ", username)
			return next(c)
		}

		c.Set("authorized_user", username)

		return next(c)
	}
}

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if _, ok := c.Get("authorized_user").(string); !ok {
			c.Response().Header().Add("HX-Redirect", "/login")
			return c.NoContent(http.StatusUnauthorized)
		}
		return next(c)
	}
}

func UserInRoomWithRoomName(db *sql.DB) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userName := c.Get("authorized_user").(string)
			roomName := c.Param("name")
			roomId, err := util.RoomExists(db, roomName)
			if err != nil {
				if err == util.ERR_ROOM_NONEXISTENT {
					return c.NoContent(http.StatusNotFound)
				}
				return err
			}
			inRoom, err := util.IsUserInRoom(db, userName, roomId)
			if !inRoom {
				return c.NoContent(http.StatusUnauthorized)
			}
			c.Set("room_id", roomId)
			return next(c)
		}
	}
}
