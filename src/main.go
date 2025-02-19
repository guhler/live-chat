package main

import (
	"database/sql"
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var (
	upgrader = websocket.Upgrader{}
	DB       *sql.DB
)

func main() {
	err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	initTempl(e)

	e.Use(middleware.Logger())
	e.Use(authMiddleware)
	e.Static("/static", "./static")

	err = routeRegister(e)
	if err != nil {
		log.Fatal(err)
	}
	err = routeLogin(e)
	if err != nil {
		log.Fatal(err)
	}

	err = routeIndex(e)
	if err != nil {
		log.Fatal(err)
	}
	err = routeLoginPage(e)
	if err != nil {
		log.Fatal(err)
	}

	err = e.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
