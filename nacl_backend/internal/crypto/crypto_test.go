package crypto

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			"returns correct length of bytes",
			32,
		},
		{
			"returns 0 length of bytes",
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testResult := GenerateRandomBytes(tt.length)

			assert.Equal(t, tt.length, len(testResult))
		})
	}
}

func TestDeriveKey(t *testing.T) {
	tests := []struct {
		name           string
		password       string
		salt           []byte
		expectError    bool
		expectedLength int
	}{
		{
			"returns correct length []byte with nil error",
			"password",
			GenerateRandomBytes(32),
			false,
			32,
		}, {
			"returns error from short salt",
			"password",
			GenerateRandomBytes(31),
			true,
			0,
		}, {
			"returns error from long salt",
			"password",
			GenerateRandomBytes(33),
			true,
			0,
		}, {
			"returns error from emtpy password",
			"",
			GenerateRandomBytes(32),
			true,
			32,
		}, {
			"returns error from nil salt",
			"password",
			nil,
			true,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testResult, err := DeriveKey(tt.password, tt.salt)
			if tt.expectError {
				assert.Error(t, err, "expected error, got %v", err)
				return
			}

			assert.Equal(t,
				tt.expectedLength,
				len(testResult),
				fmt.Sprintf("expected result to be length %d, got %d", tt.expectedLength, len(testResult)))
		})
	}

}

func TestEncrypt(t *testing.T) {
	testPassword := []byte("test_password")
	key, _ := DeriveKey("password", GenerateRandomBytes(32))
	ciphertextLength := func(password []byte) int {
		return len(password) + 28
	}

	tests := []struct {
		name           string
		toEncrypt      []byte
		key            []byte
		expectError    bool
		expectedLength int
	}{
		{
			"returns correct length ciphertext",
			testPassword,
			key,
			false,
			ciphertextLength(testPassword),
		},
		{
			"returns error from empty toEncrypt",
			nil,
			key,
			true,
			0,
		},
		{
			"returns error from short key",
			testPassword,
			GenerateRandomBytes(31),
			true,
			0,
		},
		{
			"returns error from long key",
			testPassword,
			GenerateRandomBytes(33),
			true,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testResult, err := Encrypt([]byte(tt.toEncrypt), tt.key)
			if tt.expectError {
				assert.Error(t, err, "expected error, got %v", err)
				return
			}

			assert.Equal(t,
				tt.expectedLength,
				len(testResult),
				fmt.Sprintf("expected length %d, got %d", tt.expectedLength, len(testResult)))
		})
	}

}

func TestDecrypt(t *testing.T) {
	testPassword := []byte("test_password")
	key, _ := DeriveKey("password", GenerateRandomBytes(32))
	encrypted, _ := Encrypt(testPassword, key)

	tests := []struct {
		name        string
		toDecrypt   []byte
		key         []byte
		expectError bool
		expected    []byte
	}{
		{
			"successfully decrypts ciphertext",
			encrypted,
			key,
			false,
			testPassword,
		},
		{
			"returns error from wrong key",
			encrypted,
			GenerateRandomBytes(32),
			true,
			nil,
		},
		{
			"returns error from short ciphertext",
			GenerateRandomBytes(2),
			key,
			true,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testResult, err := Decrypt(tt.toDecrypt, tt.key)
			if tt.expectError {
				assert.Error(t, err, "expected error, got %v", err)
				return
			}

			assert.Equal(t,
				tt.expected,
				testResult,
				fmt.Sprintf("expected %s, got %s", string(tt.expected), string(testResult)))
		})
	}

}
