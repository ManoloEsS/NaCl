package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/stretchr/testify/assert"
)

func TestHandleLogin(t *testing.T) {
	tests := []struct {
		name           string
		usernameCreate string
		passwordCreate string
		usernameLogin  string
		passwordLogin  string
		expectError    bool
		loginWantCode  int
	}{
		{
			"login successfull",
			"test_user",
			"password",
			"test_user",
			"password",
			false,
			200,
		},
		{
			"unsuccessful login with wrong username",
			"test_user",
			"password",
			"wrong_username",
			"password",
			true,
			401,
		},
		{
			"unsuccessful login with wrong password",
			"test_user",
			"password",
			"test_user",
			"wrong_pass",
			true,
			401,
		},
	}

	pool, queries := newTestDB(t)
	defer pool.Close()

	server := newTestServer(t, queries)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t, pool, "users")

			passHash, err := auth.HashPassword(tt.passwordCreate)
			if err != nil {
				t.Errorf("could not hash passwordCreate: %v", err)
			}
			salt, err := encryption.GenerateRandomBytes(32)
			if err != nil {
				t.Fatalf("could not generate salt: %v", err)
			}
			key, err := encryption.GenerateRandomBytes(32)
			if err != nil {
				t.Fatalf("could not generate key: %v", err)
			}
			newUserData := db.CreateUserParams{
				Username:           tt.usernameCreate,
				PasswordHash:       passHash,
				MasterKeySalt:      base64.StdEncoding.EncodeToString(salt),
				EncryptedMasterKey: base64.StdEncoding.EncodeToString(key),
			}

			err = queries.CreateUser(context.Background(), newUserData)
			if err != nil {
				t.Fatalf("insert test failed: %v", err)
			}

			body := fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, tt.usernameLogin, tt.passwordLogin)
			req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			server.HTTPServer.Handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.loginWantCode, rr.Code, "unexpected status code")
		})
	}
}
