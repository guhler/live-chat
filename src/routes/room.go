package routes

import (
	"database/sql"
	"fmt"
	"live_chat/auth"
	"live_chat/src/util"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

var (
	INITIAL_MSGS = 40
	upgrader     = websocket.Upgrader{}
)

func GetRoomPage(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name", func(c echo.Context) error {
		userName := c.Get("authorized_user").(string)

		roomName := c.Param("name")
		if roomName == "" {
			return c.NoContent(http.StatusNotFound)
		}

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
	}, auth.RequireAuth
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

func GetRoomMessages(db *sql.DB) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name/messages", func(c echo.Context) error {
		roomName := c.Param("name")
		if roomName == "" {
			return c.NoContent(http.StatusNotFound)
		}
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
	}, auth.RequireAuth
}

type messageMapKey struct {
	string
	uint64
}
type messageMapVal struct {
	ch    chan message
	close func()
}

func PostRoomMessage(db *sql.DB, chMap *sync.Map) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "POST", "/rooms/:name/messages", func(c echo.Context) error {
		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		roomName := c.Param("name")
		roomId, err := util.RoomExists(db, roomName)
		if err != nil {
			if err == util.ERR_ROOM_NONEXISTENT {
				return c.NoContent(http.StatusNotFound)
			}
			return err
		}
		inRoom, err := util.IsUserInRoom(db, userName, uint64(roomId))
		if err != nil {
			return err
		}
		if !inRoom {
			return c.NoContent(http.StatusUnauthorized)
		}

		cont := c.FormValue("message-content")
		if len(cont) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		for _, cha := range chMap.Range {
			cha.(messageMapVal).ch <- message{UserName: userName, Content: cont}
		}

		_, err = db.Exec(`
			insert into messages (user_id, room_id, time, content)
			values ((select id from users where name = ?), ?, datetime('now'), ?)
			`,
			userName, roomId, cont,
		)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusCreated)
	}, auth.RequireAuth
}

func RoomWebsocket(db *sql.DB, chMap *sync.Map) (string, string, echo.HandlerFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name/messages/ws", func(c echo.Context) error {
		conHeader := c.Request().Header["Connection"]
		if len(conHeader) != 1 || conHeader[0] != "Upgrade" {
			return c.NoContent(http.StatusUpgradeRequired)
		}
		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		roomName := c.Param("name")
		roomId, err := util.RoomExists(db, roomName)
		if err != nil {
			return c.NoContent(http.StatusNotFound)
		}
		inRoom, err := util.IsUserInRoom(db, userName, roomId)
		if err != nil {
			return err
		}
		if !inRoom {
			return c.NoContent(http.StatusUnauthorized)
		}

		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}
		defer ws.Close()

		cookie, err := c.Cookie("auth-token")
		if err != nil {
			return err
		}

		cha := make(chan message)
		tk := cookie.Value
		oldChan, ok := chMap.Swap(messageMapKey{tk, roomId}, messageMapVal{cha, sync.OnceFunc(func() { close(cha) })})
		if ok {
			oldChan.(messageMapVal).close()
		}

		for msg := range cha {
			wr, err := ws.NextWriter(websocket.TextMessage)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					oldChan, ok := chMap.LoadAndDelete(messageMapKey{tk, roomId})
					if ok {
						oldChan.(messageMapVal).close()
					}
				}
				return err
			}
			err = c.Echo().Renderer.Render(
				wr,
				"index_auth/ws-message",
				msg,
				c,
			)
			wr.Close()
			if err != nil {
				return err
			}
		}
		return nil
	}, auth.RequireAuth
}
