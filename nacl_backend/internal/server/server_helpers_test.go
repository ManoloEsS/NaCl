package server

import (
	"log"
	"testing"

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
			d := newTestDB(t)
			defer d.Close()
			s := newTestServer(t, d)

			hashedPassword, err := s.hashPassword(tt.password)
			log.Println(hashedPassword)
			assert.NoError(t, err, "expected err to be nil, got %v", err)

			if tt.expectedErr == true {
				match, _ := s.checkPasswordHash("wrongpass", hashedPassword)
				assert.False(t, match, "expected false match")
				return
			}

			match, err := s.checkPasswordHash(tt.password, hashedPassword)
			assert.True(t, match, "expected match to be true")
		})
	}
}
