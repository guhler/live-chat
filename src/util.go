package main

func validateUserName(name string) int {
	for i, c := range name {
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

func validatePassword(password string) int {
	for i := 0; i < len(password); i++ {
		if password[i] <= 32 || password[i] >= 127 {
			return i
		}
	}
	return -1
}
