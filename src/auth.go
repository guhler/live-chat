package main

import (
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
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

func route_register(e *echo.Echo) error {

	e.POST("/auth/register", func(c echo.Context) error {
		regReq := registerRequest{}
		err := c.Bind(&regReq)
		if err != nil {
			return c.JSON(
				http.StatusBadRequest,
				map[string]any{"error": "Bad Request"},
			)
		}

		err = addUser(DB, regReq.Name, regReq.Password)
		if err != nil {
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

func route_login(e *echo.Echo) error {

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
			header := c.Response().Header()
			header["Auth-Token"] = []string{"1234"}
			return c.HTML(http.StatusOK, "")
		}
		return nil
	})

	return nil
}

func genToken(username string) (string, error) {
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "live_chat",
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(token_expiry_time)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	return tk.SignedString(jwtSecret)
}

func validateToken(tkString string) (string, error) {
	tk, err := jwt.Parse(tkString, func(tk *jwt.Token) (any, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := tk.Claims.(*jwt.RegisteredClaims); ok && tk.Valid {
		return claims.Subject, nil
	}
	return "", nil
}
