package routes

import (
	"database/sql"
	"fmt"
	"live_chat/auth"
	"live_chat/templ"
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

func GetRoomsPage(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms", func(c echo.Context) error {
		userName := c.Get("authorized_user").(string)
		roomsStr, err := util.GetRoomsOfUser(db, userName)
		if err != nil {
			return err
		}

		page := make(templ.RoomsPage, len(roomsStr))
		for i, s := range roomsStr {
			page[i] = struct{ RoomName string }{s}
		}

		return templ.RenderRoomsPage(c, http.StatusOK, page)
	}, auth.RequireAuth
}

func GetRoomPage(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name", func(c echo.Context) error {
		userName := c.Get("authorized_user").(string)
		roomName := c.Param("name")

		roomNames, err := util.GetRoomsOfUser(db, userName)
		if err != nil {
			return err
		}
		sidebar := make(templ.Sidebar, len(roomNames))
		for i, s := range roomNames {
			sidebar[i] = templ.RoomButton{RoomName: s, Selected: s == roomName}
		}

		msgs, err := util.GetMessages(db, roomName, 0, int64(INITIAL_MSGS))
		if err != nil {
			return err
		}
		messages := make([]templ.Message, len(msgs))
		for i, s := range msgs {
			messages[i] = templ.Message{UserName: s[0], Content: s[1], IsOwn: s[0] == userName}
		}
		return templ.RenderRoomPage(c, http.StatusOK, templ.RoomPage{
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
			return templ.RenderNewRoomError(c, http.StatusBadRequest, err.Error())
		}

		roomId, err := util.AddRoom(db, roomName)
		if err != nil {
			if sqliteErr, ok := err.(sqlite3.Error); ok &&
				sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return templ.RenderNewRoomError(c, http.StatusConflict, "room already exists")
			}
			return err
		}

		err = util.AddUserToRoom(db, roomId, userName)
		if err != nil {
			return err
		}

		return templ.RenderRoomBtn(c,
			http.StatusCreated,
			templ.RoomButton{
				RoomName:   roomName,
				Selected:   false,
				CurrentUrl: c.Request().Header.Get("HX-Current-Url"),
			},
		)
	}, auth.RequireAuth
}

func GetRoomMessages(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name/messages", func(c echo.Context) error {
		userName := c.Get("authorized_user")
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
		messages := make([]templ.Message, len(msgs))
		for i, s := range msgs {
			messages[i] = templ.Message{UserName: s[0], Content: s[1], IsOwn: s[0] == userName}
		}

		if c.QueryParam("initial") == "true" {
			return templ.RenderSwitchRoom(c, http.StatusOK, templ.SwitchRoom{
				RoomName: roomName,
				ChatContent: templ.MessageList{
					RoomName:  roomName,
					Messages:  messages,
					Done:      len(messages) < count,
					NextStart: start + count,
				},
			})
		}
		return templ.RenderMessageList(c, http.StatusOK, templ.MessageList{
			RoomName:  roomName,
			Messages:  messages,
			Done:      len(messages) < count,
			NextStart: start + count,
		})
	}, auth.RequireAuth, auth.UserInRoomWithRoomName(db)
}
