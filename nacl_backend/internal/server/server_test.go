package server

import (
	"testing"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/stretchr/testify/assert"
)

func TestPasswordHash(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr bool
	}{
		{
			name:        "password matches",
			password:    "password",
			expectedErr: false,
		},
		{
			name:        "empty password matches",
			password:    "",
			expectedErr: false,
		},
		{
			name:        "password doesn't match hash",
			password:    "will_not_match",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := auth.HashPassword(tt.password)
			assert.NoError(t, err, "expected err to be nil, got %v", err)

			if tt.expectedErr == true {
				match, _ := auth.CheckPasswordHash("wrongpass", hashedPassword)
				assert.False(t, match, "expected false match")
				return
			}

			match, _ := auth.CheckPasswordHash(tt.password, hashedPassword)
			assert.True(t, match, "expected match to be true")
		})
	}
}

// TODO: add unittest for decodeAndValidate function
