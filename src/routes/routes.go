package routes

import (
	"database/sql"
	"sync"

	"github.com/labstack/echo/v4"
)

func AddAll(e *echo.Echo, db *sql.DB) {

	e.Add(GetIndex(db))

	e.Add(GetRoomPage(db))
	e.Add(GetRoomMessages(db))
	e.Add(PostRoom(db))

	e.Add(GetLoginPage())
	e.Add(Login(db))
	e.Add(GetRegisterPage())
	e.Add(Register(db))

	mp := sync.Map{}
	e.Add(RoomWebsocket(db, &mp))
}
