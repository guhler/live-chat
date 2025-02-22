package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

var (
	jwtSecret []byte

	token_expiry_time = time.Hour * 4
)

func initAuth() error {
	sec := os.Getenv("JWT_SECRET")
	if sec == "" {
		return errors.New("JWT_SECRET not provided in environment")
	}
	jwtSecret = []byte(sec)
	return nil
}

type registerRequest struct {
	Name     string `form:"username"`
	Password string `form:"password"`
}

func routeRegister(e *echo.Echo) error {

	e.POST("/register", func(c echo.Context) error {
		regReq := registerRequest{}
		err := c.Bind(&regReq)
		if err != nil {
			return c.HTML(
				http.StatusBadRequest,
				"Bad Request",
			)
		}

		if i := validateUserName(regReq.Name); i != -1 {
			bad_char := string(regReq.Name[i])
			if bad_char == " " {
				bad_char = "spaces"
			}
			return c.HTML(
				http.StatusBadRequest,
				"User name cannot contain "+bad_char,
			)
		}
		if i := validatePassword(regReq.Password); i != -1 {
			bad_char := string(regReq.Name[i])
			if bad_char == " " {
				bad_char = "spaces"
			}
			return c.HTML(
				http.StatusBadRequest,
				"Password cannot contain "+bad_char,
			)
		}

		err = addUser(DB, regReq.Name, regReq.Password)
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

		tk, err := genToken(regReq.Name)
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
	})

	return nil
}

type loginRequest struct {
	Name     string `form:"username"`
	Password string `form:"password"`
}

func routeLogin(e *echo.Echo) error {

	e.POST("/login", func(c echo.Context) error {
		logReq := loginRequest{}
		err := c.Bind(&logReq)
		if err != nil {
			return err
		}

		if i := validateUserName(logReq.Name); i != -1 {
			return c.Render(
				http.StatusNotFound,
				"login/error",
				"User does not exist",
			)
		}
		if i := validatePassword(logReq.Password); i != -1 {
			return c.Render(
				http.StatusUnauthorized,
				"login/error",
				"Invalid password",
			)
		}

		switch validateCredentials(DB, logReq.Name, logReq.Password) {
		case USER_DOES_NOT_EXIST:
			return c.Render(
				http.StatusNotFound,
				"login/error",
				"User does not exist",
			)
		case INVALID_PASSWORD:
			return c.Render(
				http.StatusUnauthorized,
				"login/error",
				"Invalid password",
			)
		case OK:
			tk, err := genToken(logReq.Name)
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
	})

	return nil
}

func routeLogout(e *echo.Echo) error {
	e.POST("/logout", func(c echo.Context) error {
		un := c.Get("authorized_user")
		if un == nil {
			return c.HTML(
				http.StatusUnauthorized,
				"Not logged in",
			)
		}
		username := un.(string)

		err := logoutUser(DB, username)
		if err != nil {
			return err
		}

		c.Response().Header().Add("HX-Redirect", "/")
		c.SetCookie(&http.Cookie{
			Name:   "auth-token",
			Path:   "/",
			MaxAge: -1,
		})
		return c.HTML(
			http.StatusCreated,
			"Logged out",
		)
	})
	return nil
}

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("auth-token")
		if err != nil {
			if err != echo.ErrCookieNotFound {
				c.Logger().Warn("Cookie Error: ", err)
			}
			return next(c)
		}

		tk := cookie.Value
		username, err := validateToken(tk)
		if err != nil {
			c.Logger().Warn("Failed to validate Token: ", err)
			return next(c)
		}
		if validateUserName(username) != -1 {
			c.Logger().Warn("Parsed invalid user name from Token: ", username)
			return next(c)
		}

		c.Set("authorized_user", username)

		return next(c)
	}
}

func genToken(username string) (string, error) {
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "live_chat",
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(token_expiry_time)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
	})

	return tk.SignedString(jwtSecret)
}

func validateToken(tkString string) (string, error) {
	tk, err := jwt.ParseWithClaims(tkString, &jwt.RegisteredClaims{}, func(tk *jwt.Token) (any, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return "", err
	}

	if claims, ok := tk.Claims.(*jwt.RegisteredClaims); ok && tk.Valid {
		row := DB.QueryRow("select logout_time from users where name = ?", claims.Subject)
		var logout_time time.Time
		err := row.Scan(&logout_time)
		if err != nil {
			return "", err
		}
		// if logged out after token generation
		if logout_time.Compare(claims.IssuedAt.Time) == 1 {
			return "", nil
		}
		return claims.Subject, nil
	}
	return "", nil
}
