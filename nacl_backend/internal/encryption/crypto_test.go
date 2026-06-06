package encryption

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

const testPassword = "password"

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
			testPassword,
			GenerateRandomBytes(32),
			false,
			32,
		}, {
			"returns error from short salt",
			testPassword,
			GenerateRandomBytes(31),
			true,
			0,
		}, {
			"returns error from long salt",
			testPassword,
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
			testPassword,
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
	testPayload := []byte("test_password")
	key, _ := DeriveKey(testPassword, GenerateRandomBytes(32))
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
			testPayload,
			key,
			false,
			ciphertextLength(testPayload),
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
			testPayload,
			GenerateRandomBytes(31),
			true,
			0,
		},
		{
			"returns error from long key",
			testPayload,
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
	testPayload := []byte("test_password")
	key, _ := DeriveKey(testPassword, GenerateRandomBytes(32))
	encrypted, _ := Encrypt(testPayload, key)

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
			testPayload,
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
