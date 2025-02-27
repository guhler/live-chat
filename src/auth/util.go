package auth

import (
	"database/sql"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

func GenToken(username string) (string, error) {
	tk := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "live_chat",
		Subject:   username,
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(token_expiry_time)),
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
	})

	return tk.SignedString(jwtSecret)
}

func ValidateToken(tkString string) (string, error) {
	tk, err := jwt.ParseWithClaims(tkString, &jwt.RegisteredClaims{}, func(tk *jwt.Token) (any, error) {
		return jwtSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		return "", err
	}

	if claims, ok := tk.Claims.(*jwt.RegisteredClaims); ok && tk.Valid {
		row := DB.QueryRow("select logout_time from users where name = ?", claims.Subject)
		var logout_time time.Time
		err := row.Scan(&logout_time)
		if err != nil {
			return "", err
		}
		// if logged out after token generation
		if logout_time.Compare(claims.IssuedAt.Time) == 1 {
			return "", nil
		}
		return claims.Subject, nil
	}
	return "", nil
}
