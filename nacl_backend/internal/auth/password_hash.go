package auth

import "github.com/alexedwards/argon2id"

// HashPassword generates a secure Argon2id hash of the provided password.
// Used by HandlerCreateUser and HandlerUserUpdate to store passwords securely.
func HashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// CheckPasswordHash compares a plaintext password with an Argon2id hash.
// Used by HandlerUserLogin to verify user credentials during authentication.
func CheckPasswordHash(password, hash string) (bool, error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return match, err
	}

	return match, nil
}
