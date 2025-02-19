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

	e.Use(middleware.Logger())
	e.Static("", "./static")

	err = route_register(e)
	if err != nil {
		log.Fatal(err)
	}
	err = route_login(e)
	if err != nil {
		log.Fatal(err)
	}

	err = e.Start(":8080")

	if err != nil {
		log.Fatal(err)
	}
}
