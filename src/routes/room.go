package routes

import (
	"database/sql"
	"fmt"
	"live_chat/auth"
	"live_chat/util"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

var (
	INITIAL_MSGS = 40
	upgrader     = websocket.Upgrader{}
)

func GetRoomPage(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name", func(c echo.Context) error {
		userName := c.Get("authorized_user").(string)
		roomName := c.Param("name")

		roomNames, err := util.GetRoomsOfUser(db, userName)
		if err != nil {
			return err
		}
		sidebar := make(sidebar, len(roomNames))
		for i, s := range roomNames {
			sidebar[i] = roomButton{RoomName: s, Selected: s == roomName}
		}

		msgs, err := util.GetMessages(db, roomName, 0, int64(INITIAL_MSGS))
		if err != nil {
			return err
		}
		messages := make([]message, len(msgs))
		for i, s := range msgs {
			messages[i] = message{s[0], s[1]}
		}
		return c.Render(http.StatusOK, "room.html", roomPage{
			RoomName:  roomName,
			Sidebar:   sidebar,
			WsUrl:     fmt.Sprintf("/rooms/%s/messages/ws", roomName),
			Messages:  messages,
			Done:      len(messages) < INITIAL_MSGS,
			NextStart: INITIAL_MSGS,
		})
	}, auth.RequireAuth, auth.UserInRoomWithRoomName(db)
}

func PostRoom(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "POST", "/rooms", func(c echo.Context) error {

		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.String(http.StatusUnauthorized, "Not authorized")
		}

		roomName := c.FormValue("room-name")

		if err := util.ValidateRoomName(roomName); err != nil {
			return c.Render(http.StatusBadRequest, "index_auth/new-room-error", err.Error())
		}

		roomId, err := util.AddRoom(db, roomName)
		if err != nil {
			if sqliteErr, ok := err.(sqlite3.Error); ok &&
				sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return c.Render(http.StatusConflict, "index_auth/new-room-error", "Room already exists")
			}
			return err
		}

		err = util.AddUserToRoom(db, roomId, userName)
		if err != nil {
			return err
		}
		return c.Render(http.StatusCreated, "room/post-response", roomButton{RoomName: roomName, Selected: false})
	}, auth.RequireAuth
}

func GetRoomMessages(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name/messages", func(c echo.Context) error {
		roomName := c.Param("name")

		start, err := strconv.Atoi(c.QueryParam("start"))
		if err != nil {
			start = 0
		}
		count, err := strconv.Atoi(c.QueryParam("count"))
		if err != nil {
			count = 20
		}

		msgs, err := util.GetMessages(db, roomName, int64(start), int64(count))
		if err != nil {
			return err
		}
		messages := make([]message, len(msgs))
		for i, s := range msgs {
			messages[i] = message{UserName: s[0], Content: s[1]}
		}

		fmt.Println(c.Request())

		return c.Render(http.StatusOK, "room/message-response", messageResponse{
			RoomName:         roomName,
			Selected:         true,
			SelectedRoomName: "",
			ChatContent: chatContent{
				RoomName:  roomName,
				Messages:  messages,
				Done:      len(messages) < count,
				NextStart: start + count,
			},
		})
	}, auth.RequireAuth, auth.UserInRoomWithRoomName(db)
}
