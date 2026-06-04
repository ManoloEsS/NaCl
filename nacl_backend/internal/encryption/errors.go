package encryption

import "errors"

var (
	ErrInvalidLengthPassword = errors.New("password length cannot be 0")
	ErrInvalidLengthSalt     = errors.New("length of salt must be 32 bytes")
	ErrInvalidLengthKey      = errors.New("length of key must be 32 bytes")
	ErrEmptyPassword         = errors.New("password cannot be empty")
	ErrEmptyPlaintext        = errors.New("plaintext cannot be empty")
	ErrCiphertextShort       = errors.New("ciphertext too short")
)
