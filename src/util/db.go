package util

import (
	"database/sql"
	"errors"
)

func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "file:../db.sqlite")
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func LogoutUser(db *sql.DB, name string) error {
	_, err := db.Exec("update users set logout_time = datetime('now') where name = ?", name)
	return err
}

func AddRoom(db *sql.DB, name string) (roomId int64, err error) {
	res, err := db.Exec("insert into rooms (name) values (?)", name)
	if err != nil {
		return
	}
	roomId, err = res.LastInsertId()
	return
}

func AddUserToRoom(db *sql.DB, roomId int64, userName string) error {
	_, err := db.Exec(`
		insert into room_user (user_id, room_id) 
		values ((select id from users where name = ?), ?)`,
		userName, roomId,
	)
	return err
}

func IsUserInRoom(db *sql.DB, userName string, roomId uint64) (bool, error) {
	rows, err := db.Query(`
		select null from room_user
		where room_id = ? and user_id = (select id from users where name = ?)`,
		roomId, userName,
	)
	defer rows.Close()
	if err != nil {
		return false, err
	}
	return rows.Next(), nil
}

var ERR_ROOM_NONEXISTENT = errors.New("room does not exist")

func RoomExists(db *sql.DB, roomName string) (uint64, error) {
	rows, err := db.Query(`select id from rooms where name = ?`, roomName)
	defer rows.Close()
	if err != nil {
		return 0, err
	}
	if !rows.Next() {
		return 0, ERR_ROOM_NONEXISTENT
	}
	var id uint64
	rows.Scan(&id)
	return id, nil
}

func GetMessages(db *sql.DB, roomName string, start, count int64) ([][2]string, error) {
	rows, err := db.Query(`
		select users.name, messages.content
		from messages
		join users on messages.user_id = users.id
		where messages.room_id = (select id from rooms where name = ?)
		order by messages.time desc
		limit ? offset ?`,
		roomName, count, start,
	)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	messages := [][2]string{}
	for rows.Next() {
		var userName, content string
		err := rows.Scan(&userName, &content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, [2]string{userName, content})
	}
	return messages, nil
}

func GetRoomsOfUser(db *sql.DB, userName string) ([]string, error) {
	rows, err := db.Query(`
			select rooms.name
			from rooms
			join room_user on rooms.id = room_user.room_id
			join users on users.id = room_user.user_id
			where users.name = ?`,
		userName,
	)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	roomNames := []string{}
	for rows.Next() {
		var room string
		err := rows.Scan(&room)
		if err != nil {
			return nil, err
		}
		roomNames = append(roomNames, room)
	}
	return roomNames, nil
}
