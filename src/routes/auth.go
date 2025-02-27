package routes

import (
	"database/sql"
	"live_chat/auth"
	"live_chat/util"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

func GetLoginPage() (string, string, echo.HandlerFunc) {
	return "GET", "/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	}
}

func GetRegisterPage() (string, string, echo.HandlerFunc) {
	return "GET", "/register", func(c echo.Context) error {
		return c.Render(http.StatusOK, "register.html", nil)
	}
}

func Register(db *sql.DB) (string, string, echo.HandlerFunc) {
	return "POST", "/register", func(c echo.Context) error {
		var regReq struct {
			Name     string `form:"username"`
			Password string `form:"password"`
		}
		err := c.Bind(&regReq)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		if err := util.ValidateUserName(regReq.Name); err != nil {
			return c.HTML(
				http.StatusBadRequest,
				err.Error(),
			)
		}
		if err := util.ValidatePassword(regReq.Password); err != nil {
			return c.HTML(
				http.StatusBadRequest,
				err.Error(),
			)
		}

		err = util.AddUser(db, regReq.Name, regReq.Password)
		if err != nil {
			if sqliteErr, ok := err.(sqlite3.Error); ok &&
				sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return c.HTML(
					http.StatusConflict,
					"User name already exists",
				)
			}
			return err
		}

		tk, err := auth.GenToken(regReq.Name)
		if err != nil {
			return err
		}
		c.SetCookie(&http.Cookie{
			Name:     "auth-token",
			Value:    tk,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
		})
		c.Response().Header().Add("HX-Redirect", "/")

		return c.HTML(
			http.StatusCreated,
			"User created",
		)
	}
}

func Login(db *sql.DB) (string, string, echo.HandlerFunc) {
	return "POST", "/login", func(c echo.Context) error {
		var logReq struct {
			Name     string `form:"username"`
			Password string `form:"password"`
		}
		err := c.Bind(&logReq)
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		if util.ValidateUserName(logReq.Name) != nil {
			return c.Render(
				http.StatusNotFound,
				"login/error",
				"User does not exist",
			)
		}
		if util.ValidatePassword(logReq.Password) != nil {
			return c.Render(
				http.StatusUnauthorized,
				"login/error",
				"Invalid password",
			)
		}

		switch util.ValidateCredentials(db, logReq.Name, logReq.Password) {
		case util.USER_DOES_NOT_EXIST:
			return c.Render(
				http.StatusNotFound,
				"login/error",
				"User does not exist",
			)
		case util.INVALID_PASSWORD:
			return c.Render(
				http.StatusUnauthorized,
				"login/error",
				"Invalid password",
			)
		case util.OK:
			tk, err := auth.GenToken(logReq.Name)
			if err != nil {
				return err
			}
			c.SetCookie(&http.Cookie{
				Name:     "auth-token",
				Value:    tk,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
				SameSite: http.SameSiteStrictMode,
			})
			c.Response().Header().Add("HX-Redirect", "/")
			return c.String(http.StatusOK, "")
		}
		return nil
	}
}

func Logout(db *sql.DB) (string, string, echo.HandlerFunc) {
	return "POST", "/logout", func(c echo.Context) error {
		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.HTML(
				http.StatusUnauthorized,
				"Not logged in",
			)
		}

		err := util.LogoutUser(db, userName)
		if err != nil {
			return err
		}

		c.Response().Header().Add("HX-Redirect", "/")
		c.SetCookie(&http.Cookie{
			Name:   "auth-token",
			Path:   "/",
			MaxAge: -1,
		})
		return c.NoContent(http.StatusCreated)
	}
}
