package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	tests := []struct {
		name         string
		userID       uuid.UUID
		tokenSecret  string
		expiresIn    time.Duration
		validateWith string
		sleepTime    time.Duration
		expectError  bool
	}{
		{
			name:         "valid",
			userID:       uuid.New(),
			tokenSecret:  "test",
			expiresIn:    time.Duration(time.Second * 30),
			validateWith: "test",
			sleepTime:    time.Millisecond * 1,
			expectError:  false,
		},
		{
			name:         "invalid",
			userID:       uuid.New(),
			tokenSecret:  "test",
			expiresIn:    time.Duration(time.Second * 30),
			validateWith: "wrong",
			sleepTime:    time.Millisecond * 1,
			expectError:  true,
		},
		{
			name:        "expired",
			userID:      uuid.New(),
			tokenSecret: "test",
			expiresIn:   time.Duration(time.Millisecond * 5),
			sleepTime:   time.Millisecond * 20,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwt, err := MakeJWT(tt.userID, tt.tokenSecret, tt.expiresIn)
			require.NoError(t, err, "MakeJWT should succeed")

			time.Sleep(tt.sleepTime)
			validatedID, err := ValidateJWT(jwt, tt.validateWith)

			if tt.expectError {
				assert.Error(t, err, "expected validation to fail")
			} else {
				assert.NoError(t, err, "expected validation to succeed")
				assert.Equal(t, tt.userID, validatedID, "User ID should match")
			}
		})
	}
}
