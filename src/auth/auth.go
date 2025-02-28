package auth

import (
	"database/sql"
	"time"
)

var (
	jwtSecret         []byte
	token_expiry_time = time.Hour * 4
	DB                *sql.DB
)

func Init(secret []byte, db *sql.DB) {
	jwtSecret = secret
	DB = db
}
