package main

import (
	"database/sql"
)

func initDB() error {
	db, err := sql.Open("sqlite3", "file:./db.sqlite")
	if err != nil {
		return err
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func addUser(db *sql.DB, name string, password string) error {
	_, err := db.Exec("insert into users (name, password) values (?, ?)", name, password)
	return err
}

func logoutUser(db *sql.DB, name string) error {
	_, err := db.Exec("update users set logout_time = datetime('now') where name = ?", name)
	return err
}

func addRoom(db *sql.DB, name string) (roomId int64, err error) {
	res, err := db.Exec("insert into rooms (name) values (?)", name)
	if err != nil {
		return
	}
	roomId, err = res.LastInsertId()
	return
}

func addUserToRoom(db *sql.DB, roomId int64, userName string) error {
	_, err := db.Exec(`
		insert into room_user (user_id, room_id) 
		values ((select id from users where name = ?), ?)`,
		userName, roomId,
	)
	return err
}

func isUserInRoom(db *sql.DB, userName string, roomId int64) (bool, error) {
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

const (
	USER_DOES_NOT_EXIST = iota
	INVALID_PASSWORD
	OK
)

func validateCredentials(db *sql.DB, name, password string) int {
	row := db.QueryRow("select password from users where name = ?", name)
	passwd := ""
	err := row.Scan(&passwd)
	if err != nil {
		return USER_DOES_NOT_EXIST
	}
	if password != passwd {
		return INVALID_PASSWORD
	}
	return OK
}
