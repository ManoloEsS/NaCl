package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"golang.org/x/crypto/argon2"
)

func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return b
}

func DeriveKey(password string, salt []byte) ([]byte, error) {
	if len(salt) != 32 {
		return nil, fmt.Errorf("salt must be 32 bytes, got %d", len(salt))
	}

	if len(password) == 0 {
		return nil, fmt.Errorf("password cannot be empty string")
	}
	derivedKey := argon2.Key([]byte(password), salt, 3, 32*1024, 4, 32)

	return derivedKey, nil
}

// caller can get string with -> base64.StdEncoding.EncodeToString(returnedCiphertext)
func Encrypt(toEncrypt, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
	}
	if len(toEncrypt) == 0 {
		return nil, fmt.Errorf("bytes to array can't be empty")
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

// caller can get string with -> string(returnedPlaintext)
func Decrypt(ciphertext, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes, got %d", len(key))
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
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	decrypted, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return decrypted, nil
}
