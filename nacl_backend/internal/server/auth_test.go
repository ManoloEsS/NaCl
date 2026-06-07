package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestHandlerLogin(t *testing.T) {
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

	testDB := newTestDB(t)
	defer testDB.Close()

	server := newTestServer(t, testDB)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t, testDB, "users")

			passHash, err := auth.HashPassword(tt.passwordCreate)
			if err != nil {
				t.Errorf("could not hash passwordCreate: %v", err)
			}
			newUserData := db.CreateUserParams{
				Username:           tt.usernameCreate,
				PasswordHash:       passHash,
				MasterKeySalt:      base64.StdEncoding.EncodeToString(encryption.GenerateRandomBytes(32)),
				EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryption.GenerateRandomBytes(32)),
			}

			queries := testDB.Queries()
			user, err := queries.CreateUser(context.Background(), newUserData)
			if err != nil {
				t.Fatalf("insert test failed: %v", err)
			}

			body := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, tt.usernameLogin, tt.passwordLogin)
			req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			server.handlerLogin(rr, req)

			var userDataLogin LoginResponse
			err = json.NewDecoder(rr.Body).Decode(&userDataLogin)
			if err != nil {
				t.Errorf("could not decode user data from login: %v", err)
			}

			assert.Equal(t, tt.loginWantCode, rr.Code, "unexpected status code")
			assert.IsType(t, uuid.UUID{}, userDataLogin.ID, "login response does not contain UUID id")

			if !tt.expectError {
				assert.Equal(t, user.Username, userDataLogin.Username)
				assert.NotEqual(t, uuid.Nil, userDataLogin.ID, "expected id to be not nil")
				return
			}

		})
	}
}
