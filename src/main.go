package main

import (
	"live_chat/auth"
	"live_chat/routes"
	"live_chat/util"
	"log"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	DB, err := util.InitDB()
	if err != nil {
		log.Fatal(err)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("please provide JWT_SECRET in env")
	}
	auth.Init([]byte(secret), DB)

	e := echo.New()

	initTempl(e)

	e.Use(middleware.Logger())
	e.Use(auth.TokenParser)
	e.Static("/static", "../static")

	routes.AddAll(e, DB)

	err = e.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
