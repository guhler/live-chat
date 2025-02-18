package main

import (
	"log"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var (
	upgrader = websocket.Upgrader{}
)

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Static("", "./static")

	e.GET("/ws", websock())

	err := e.Start(":8080")

	if err != nil {
		log.Fatal(err)
	}
}

func websock() func(echo.Context) error {

	history := make([]string, 0)
	hist_recv := make(chan string)
	chans := map[chan string]bool{hist_recv: false}

	go func() {
		for m := range hist_recv {
			history = append(history, m)
		}
	}()

	return func(c echo.Context) error {
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		for _, m := range history {
			ws.WriteMessage(websocket.TextMessage, []byte(m))
		}

		ch := make(chan string)
		chans[ch] = false

		go func() {
			for {
				_, msg, err := ws.ReadMessage()
				if err != nil {
					if _, ok := err.(*websocket.CloseError); ok {
						ws.Close()
						delete(chans, ch)
						close(ch)
					} else {
						c.Logger().Error(err)
					}
					break
				}
				for channel := range chans {
					channel <- string(msg)
				}
			}
		}()

		go func() {
			for m := range ch {
				err := ws.WriteMessage(websocket.TextMessage, []byte(m))
				if err != nil {
					c.Logger().Error(err)
					break
				}
			}
		}()

		return nil
	}
}
