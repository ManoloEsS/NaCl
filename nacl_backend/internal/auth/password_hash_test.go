package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	// TODO: hash "password", assert no error, assert hash starts with "$argon2id$"
	// TODO: hash "", assert no error
}

func TestCheckPasswordHash(t *testing.T) {
	// TODO: hash "password", check same password → match
	// TODO: hash "password", check "wrong" → no match
	// TODO: hash "", check "" → match
}
