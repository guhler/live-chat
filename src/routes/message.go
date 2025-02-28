package routes

import (
	"database/sql"
	"live_chat/auth"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type messageMapKey struct {
	string
	uint64
}
type messageMapVal struct {
	ch    chan message
	close func()
}

func PostRoomMessage(db *sql.DB, chMap *sync.Map) (string, string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
	return "POST", "/rooms/:name/messages", func(c echo.Context) error {
		userName := c.Get("authorized_user").(string)
		roomName := c.Param("name")

		cont := c.FormValue("message-content")
		if len(cont) == 0 {
			return c.NoContent(http.StatusBadRequest)
		}

		for _, cha := range chMap.Range {
			cha.(messageMapVal).ch <- message{UserName: userName, Content: cont}
		}

		_, err := db.Exec(`
			insert into messages (user_id, room_id, time, content)
			values (
				(select id from users where name = ?),
				(select id from rooms where name = ?),
				datetime('now'), ?
			)
			`,
			userName, roomName, cont,
		)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusCreated)
	}, auth.RequireAuth, auth.UserInRoomWithRoomName(db)
}

func RoomWebsocket(db *sql.DB, chMap *sync.Map) (string, string, echo.HandlerFunc, echo.MiddlewareFunc, echo.MiddlewareFunc) {
	return "GET", "/rooms/:name/messages/ws", func(c echo.Context) error {
		conHeader := c.Request().Header["Connection"]
		if len(conHeader) != 1 || conHeader[0] != "Upgrade" {
			return c.NoContent(http.StatusUpgradeRequired)
		}

		userName := c.Get("authorized_user")
		roomId := c.Get("room_id").(uint64)

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
			msg.IsOwn = msg.UserName == userName
			wr, err := ws.NextWriter(websocket.TextMessage)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err) {
					oldChan, ok := chMap.LoadAndDelete(messageMapKey{tk, roomId})
					if ok {
						oldChan.(messageMapVal).close()
					}
					return nil
				}
				return err
			}
			err = c.Echo().Renderer.Render(
				wr,
				"room/ws-message",
				msg,
				c,
			)
			wr.Close()
			if err != nil {
				return err
			}
		}
		return nil
	}, auth.RequireAuth, auth.UserInRoomWithRoomName(db)
}
