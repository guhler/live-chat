package main

import (
	"database/sql"
	"fmt"
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
	// WARN: sql injection, use prepared statement, name and password could contain quotes
	_, err := db.Exec(fmt.Sprintf("insert into users (name, password) values ('%s', '%s')", name, password))
	return err
}

func logoutUser(db *sql.DB, name string) error {
	_, err := db.Exec(fmt.Sprintf("update users set logout_time = datetime('now') where name = '%s'", name))
	return err
}

func addRoom(db *sql.DB, name string) error {
	_, err := db.Exec(fmt.Sprintf("insert into rooms (name) values ('%s')", name))
	return err
}

func addUserToRoom(db *sql.DB, room_id int64, user_id int64) error {
	_, err := db.Exec(fmt.Sprintf("insert into room_user (user_id, room_id) values ('%d', '%d')", user_id, room_id))
	return err
}

const (
	USER_DOES_NOT_EXIST = iota
	INVALID_PASSWORD
	OK
)

func validateCredentials(db *sql.DB, name, password string) int {
	row := db.QueryRow(fmt.Sprintf("select password from users where name = '%s'", name))
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
