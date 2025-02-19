package main

import (
	"errors"
	"fmt"
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
			return c.JSON(
				http.StatusBadRequest,
				map[string]any{"error": "Bad Request"},
			)
		}

		if i := validateUserName(regReq.Name); i != -1 {
			return c.JSON(
				http.StatusBadRequest,
				// TODO: add special case for spaces
				map[string]any{"error": "User name cannot contain " + string(regReq.Name[i])},
			)
		}
		if i := validatePassword(regReq.Password); i != -1 {
			return c.JSON(
				http.StatusBadRequest,
				// TODO: add special case for spaces
				map[string]any{"error": "Password cannot contain " + string(regReq.Password[i])},
			)
		}

		err = addUser(DB, regReq.Name, regReq.Password)
		if err != nil {
			if sqliteErr, ok := err.(sqlite3.Error); ok &&
				sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return c.JSON(
					http.StatusConflict,
					map[string]any{"error": "User name already exists"},
				)
			}
			return err
		}

		return c.JSON(
			http.StatusCreated,
			map[string]any{},
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
			return c.JSON(
				http.StatusBadRequest,
				map[string]any{"error": "User name cannot contain " + string(logReq.Name[i])},
			)
		}
		if i := validatePassword(logReq.Password); i != -1 {
			return c.JSON(
				http.StatusBadRequest,
				map[string]any{"error": "Password cannot contain " + string(logReq.Password[i])},
			)
		}

		switch validateCredentials(DB, logReq.Name, logReq.Password) {
		case USER_DOES_NOT_EXIST:
			return c.JSON(
				http.StatusNotFound,
				map[string]any{"error": "User does not exist"},
			)
		case INVALID_PASSWORD:
			return c.JSON(
				http.StatusUnauthorized,
				map[string]any{"error": "Invalid password"},
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
			return c.HTML(http.StatusOK, "")
		}
		return nil
	})

	return nil
}

func routeLogout(e *echo.Echo) error {
	e.POST("/logout", func(c echo.Context) error {
		un := c.Get("authorized_user")
		if un == nil {
			return c.JSON(
				http.StatusUnauthorized,
				map[string]any{"error": "Not logged in"},
			)
		}
		username := un.(string)

		err := logoutUser(DB, username)
		if err != nil {
			return err
		}

		c.Response().Header().Add("HX-Redirect", "/")
		c.SetCookie(&http.Cookie{
			Name:  "auth_token",
			Value: "",
		})
		return c.JSON(
			http.StatusCreated,
			map[string]any{"info": "Logged out"},
		)
	})
	return nil
}

func authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("auth-token")
		if err != nil {
			return next(c)
		}

		tk := cookie.Value
		username, err := validateToken(tk)
		if err != nil || username == "" {
			c.Logger().Warn("Error: ", err)
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
		ID:        "0",
		Audience:  jwt.ClaimStrings{},
		NotBefore: jwt.NewNumericDate(time.Now().UTC()),
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
		row := DB.QueryRow(fmt.Sprintf("select logout_time from users where name = '%s'", claims.Subject))
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
