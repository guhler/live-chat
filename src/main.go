package main

import (
	"database/sql"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var (
	DB *sql.DB
)

func main() {
	err := initDB()
	if err != nil {
		log.Fatal(err)
	}
	err = initAuth()
	if err != nil {
		log.Fatal(err)
	}

	e := echo.New()

	initTempl(e)

	e.Use(middleware.Logger())
	e.Use(authMiddleware)
	e.Static("/static", "./static")

	e.Add(getIndex())

	e.Add(routeLogin())
	e.Add(routeLogout())
	e.Add(routeRegister())

	e.Add(getRoomMessages())

	chanMap := make(map[chan string]bool)
	e.Add(postRoomMessage(chanMap))
	e.Add(roomWebSocket(chanMap))

	e.Add(postRooms())

	e.Add(routeLoginPage())
	e.Add(routeRegisterPage())

	err = e.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
