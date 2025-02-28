package auth

import "github.com/alexedwards/argon2id"

var (
	argonParams = argon2id.DefaultParams
)

func HashAndStoreUser(userName, password string) error {
	hash, err := argon2id.CreateHash(password, argonParams)
	if err != nil {
		return err
	}

	_, err = DB.Exec(`
		insert into users (name, password_hash)
		values (?, ?)`,
		userName,
		hash,
	)
	return err
}

func IsPasswordCorrect(userName, password string) (bool, error) {
	row := DB.QueryRow(`
		select password_hash from users
		where name = ?`,
		userName,
	)

	var hash string
	err := row.Scan(&hash)
	if err != nil {
		return false, err
	}

	ok, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		// malformed hash format
		return false, err
	}
	return ok, nil
}
