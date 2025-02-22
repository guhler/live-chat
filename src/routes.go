package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/mattn/go-sqlite3"
)

type Message struct {
	UserName string
	Content  string
}

type RoomResponse struct {
	RoomId    int64
	NextStart int
	Messages  []Message
}

type Room struct {
	ID   int64
	Name string
}

type IndexPage struct {
	Title string
	Rooms []Room
}

func getIndex(c echo.Context) error {
	user, ok := c.Get("authorized_user").(string)
	if !ok {
		return c.Render(http.StatusOK, "index.html", nil)
	}

	rows, err := DB.Query(`
			select rooms.id, rooms.name
			from rooms
			join room_user on rooms.id = room_user.room_id
			join users on users.id = room_user.user_id
			where users.name = ?
			`, user)
	if err != nil {
		return err
	}

	rooms := []Room{}
	for rows.Next() {
		var room Room
		err := rows.Scan(&room.ID, &room.Name)
		if err != nil {
			return err
		}
		rooms = append(rooms, room)
	}

	return c.Render(http.StatusOK, "index_auth.html", IndexPage{
		Title: "WIP",
		Rooms: rooms,
	})
}

type PostRoomsReq struct {
	Name string `form:"room-name"`
}

func postRooms(c echo.Context) error {

	userName, ok := c.Get("authorized_user").(string)
	if !ok {
		return c.String(http.StatusUnauthorized, "Not authorized")
	}

	var req PostRoomsReq
	err := c.Bind(&req)
	if err != nil {
		return err
	}
	roomName := req.Name

	if i := validateRoomName(roomName); i != -1 {
		if i == len(roomName) {
			return c.Render(http.StatusBadRequest, "index_auth/new-room-error", "Please provide a name")
		}

		ch := string(roomName[i])
		if ch == " " {
			ch = "spaces"
		} else {
			ch = "'" + ch + "'"
		}

		return c.Render(http.StatusBadRequest, "index_auth/new-room-error", fmt.Sprintf("Name cannot contain %s", ch))
	}

	roomId, err := addRoom(DB, roomName)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok &&
			sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return c.Render(http.StatusConflict, "index_auth/new-room-error", "Room already exists")
		}
		return err
	}

	err = addUserToRoom(DB, roomId, userName)
	if err != nil {
		return err
	}
	return c.Render(http.StatusCreated, "index_auth/room-btn-response", Room{ID: roomId, Name: roomName})
}

func getRoomMessages() (string, string, echo.HandlerFunc) {
	return "GET", "/rooms/:id/messages", func(c echo.Context) error {
		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		roomId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		inRoom, err := isUserInRoom(DB, userName, int64(roomId))
		if err != nil {
			return err
		}
		if !inRoom {
			return c.NoContent(http.StatusUnauthorized)
		}

		offset, err := strconv.Atoi(c.QueryParam("start"))
		if err != nil {
			offset = 0
		}
		limit, err := strconv.Atoi(c.QueryParam("count"))
		if err != nil {
			return c.NoContent(http.StatusBadRequest)
		}

		rows, err := DB.Query(`
			select users.name, messages.content from messages
			join users on users.id = messages.user_id
			where messages.room_id = ?
			order by messages.time desc
			limit ? offset ?`,
			roomId, limit, offset,
		)
		defer rows.Close()
		if err != nil {
			return err
		}

		messages := make([]Message, 0, limit)
		for rows.Next() {
			var name, content string
			err := rows.Scan(&name, &content)
			if err != nil {
				return err
			}
			messages = append(messages, Message{name, content})
		}

		return c.Render(http.StatusOK, "index_auth/room-response", RoomResponse{int64(roomId), offset + limit, messages})
	}
}

func postRoomMessage() (string, string, echo.HandlerFunc) {
	return "POST", "/rooms/:id/messages", func(c echo.Context) error {
		userName, ok := c.Get("authorized_user").(string)
		if !ok {
			return c.NoContent(http.StatusUnauthorized)
		}

		roomId, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}
		inRoom, err := isUserInRoom(DB, userName, int64(roomId))
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

		_, err = DB.Exec(`
			insert into messages (user_id, room_id, time, content)
			values ((select id from users where name = ?), ?, datetime('now'), ?)
			`,
			userName, roomId, cont,
		)
		if err != nil {
			return err
		}
		return c.Render(http.StatusCreated, "index_auth/message", Message{userName, cont})
	}
}
