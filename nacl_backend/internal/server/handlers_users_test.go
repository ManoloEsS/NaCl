package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerCreateUser(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantCode int
	}{
		{"success", "testuser", "password123", 201},
		{"empty username", "", "password123", 400},
		{"empty password", "testuser", "", 400},
		{"long username", strings.Repeat("a", 100), "password123", 201},
		{"special chars", "test+user@example.com", "p@$$w0rd!", 201},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			testDB := newTestDB(t)
			defer testDB.Close()

			server := newTestServer(t, testDB)

			// Create request
			body := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, tt.username, tt.password)
			req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			// Execute
			server.handlerCreateUser(rr, req)

			// Assert status code
			assert.Equal(t, tt.wantCode, rr.Code, "unexpected status code")

			// For success cases, verify response and database
			if tt.wantCode == 201 {
				assert.Contains(t, rr.Body.String(), "username")
				assert.Contains(t, rr.Body.String(), tt.username)

				// Verify in database
				ctx := context.Background()
				user, err := testDB.Queries().GetUserByUsername(ctx, tt.username)
				assert.NoError(t, err)
				assert.Equal(t, tt.username, user.Username)
				assert.NotEmpty(t, user.PasswordHash)
				assert.True(t, strings.HasPrefix(user.PasswordHash, "$argon2id$"))
			}
		})
	}
}

func TestHandlerCreateUser_Duplicate(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()

	server := newTestServer(t, testDB)

	// Create first user
	body := `{"username": "duplicate", "password": "password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.handlerCreateUser(rr, req)
	assert.Equal(t, 201, rr.Code)

	// Try to create duplicate
	req2 := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()

	server.handlerCreateUser(rr2, req2)

	// Should fail with 500 (unique constraint violation)
	assert.Equal(t, 500, rr2.Code)
}

func TestHandlerCreateUser_InvalidJSON(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()

	server := newTestServer(t, testDB)

	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(`{invalid json}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.handlerCreateUser(rr, req)

	assert.Equal(t, 400, rr.Code)
}
