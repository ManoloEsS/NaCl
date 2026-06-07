// Package encryption provides cryptographic primitives for the NaCl password manager.
//
// All functions in this package operate on raw bytes ([]byte). When storing
// cryptographic data in the database (text columns), encode to base64:
//
//	// Storage: raw bytes → base64 string
//	encodedSalt := base64.StdEncoding.EncodeToString(salt)
//	encodedKey := base64.StdEncoding.EncodeToString(encryptedMasterKey)
//
// When retrieving from the database for use in encryption functions, decode
// from base64 back to raw bytes:
//
//	// Retrieval: base64 string → raw bytes
//	salt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)
//	encryptedKey, err := base64.StdEncoding.DecodeString(user.EncryptedMasterKey)
//	derivedKey, err := DeriveKey(password, salt)
//	masterKey, err := Decrypt(encryptedKey, derivedKey)
//
// Never pass base64-encoded strings directly to encryption functions as []byte.
// []byte("MWkwRjRW...") gives you the ASCII bytes of the base64 string, not
// the original raw bytes. This will cause "invalid length" errors and silently
// produce wrong results.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Algo represents a supported encryption algorithm. Use ValidAlgorithm to
// parse a string name into an Algo value.
type Algo int

const (
	InvalidAlgo Algo = iota
	AES
)

// EncryptionAlgos maps algorithm names to their Algo values. Used by
// ValidAlgorithm to validate user input and by handlers to look up algorithms.
var EncryptionAlgos = map[string]Algo{
	"aes-gcm": AES,
}

// ValidAlgorithm checks whether an encryption algorithm name is supported.
// Returns the Algo value if found, or InvalidAlgo and an error if not.
func ValidAlgorithm(name string) (Algo, error) {
	algo, ok := EncryptionAlgos[name]
	if !ok {
		return InvalidAlgo, fmt.Errorf("algorithm not found")
	}

	return algo, nil
}

// GenerateRandomBytes returns cryptographically secure random bytes of the
// given length. Used for generating salts, nonces, and master keys.
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

// DeriveKey derives a 32-byte key from a password and salt using Argon2id.
// The salt must be exactly 32 bytes. Use base64.StdEncoding.DecodeString()
// to convert a stored salt string before passing it here.
func DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(salt) != 32 {
		return nil, ErrInvalidLengthSalt
	}

	if len(password) == 0 {
		return nil, ErrInvalidLengthPassword
	}
	derivedKey := argon2.Key([]byte(password), salt, 3, 32*1024, 4, 32)

	return derivedKey, nil
}

// Encrypt encrypts plaintext using AES-GCM with the given 32-byte key.
// Returns the nonce prepended to the ciphertext.
// Encode the result with base64.StdEncoding.EncodeToString() for database storage.
func Encrypt(toEncrypt, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidLengthKey
	}
	if len(toEncrypt) == 0 {
		return nil, ErrEmptyPlaintext
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, toEncrypt, nil)
	return ciphertext, nil

}

// Decrypt decrypts AES-GCM ciphertext using the given 32-byte key.
// Decode the stored value with base64.StdEncoding.DecodeString() before
// passing it here. The nonce must be prepended to the ciphertext
// (as returned by Encrypt).
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidLengthKey
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextShort
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
