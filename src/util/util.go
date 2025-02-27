package util

import (
	"errors"
	"fmt"
)

func ValidateUserName(name string) error {
	if len(name) < 3 {
		return errors.New("Name must be at least 3 characters long")
	}

	if m := errMsg(name, AlphaNumUnderscoreMinus(name)); m != "" {
		return errors.New("Name" + m)
	}
	return nil
}

func ValidatePassword(password string) error {
	for i := 0; i < len(password); i++ {
		if password[i] <= 32 || password[i] >= 127 {
			return errors.New("Password" + errMsg(password, i))
		}
	}
	return nil
}

func ValidateRoomName(name string) error {
	if len(name) == 0 {
		return errors.New("Provide a name")
	}
	if m := errMsg(name, AlphaNumUnderscoreMinus(name)); m != "" {
		return errors.New("Name" + m)
	}
	return nil
}

func errMsg(name string, i int) string {
	if i == -1 {
		return ""
	}
	ch := string(name[i])
	if ch == " " {
		ch = "spaces"
	} else {
		ch = "'" + ch + "'"
	}
	return fmt.Sprintf(" cannot contain %s", ch)
}

func AlphaNumUnderscoreMinus(s string) int {
	for i, c := range s {
		if c >= 'a' && c <= 'z' {
			continue
		}
		if c >= 'A' && c <= 'Z' {
			continue
		}
		if c >= '0' && c <= '9' {
			continue
		}
		if c == '-' || c == '_' {
			continue
		}
		return i
	}
	return -1
}
