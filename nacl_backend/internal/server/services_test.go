package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerCreateService(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()

	server := newTestServer(t, testDB)

	testUser := "test_services_user"
	testPass := "test_services_pass"

	body := fmt.Sprintf(`{"username": "%s", "password": "%s"}`, testUser, testPass)
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "user creation failed")

	body = fmt.Sprintf(`{"username": "%s", "password": "%s"}`, testUser, testPass)
	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "login failed")

	var loginResp LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&loginResp)
	assert.NoError(t, err, "could not decode login response")
	token := loginResp.Token

	tests := []struct {
		name                string
		service             string
		username            string
		description         string
		password            string
		encryptionAlgorithm string
		userPassword        string
		authorized          bool
		wantCode            int
	}{
		{
			"successful creation of service",
			"test_service_1",
			"test_user",
			"description",
			"service_pass_1",
			"aes-gcm",
			testPass,
			true,
			http.StatusCreated,
		},
		{
			"unauthorized - no token",
			"test_service_2",
			"test_user",
			"description",
			"service_pass_2",
			"aes-gcm",
			testPass,
			false,
			http.StatusUnauthorized,
		},
		{
			"invalid user password",
			"test_service_3",
			"test_user",
			"description",
			"service_pass_3",
			"aes-gcm",
			"wrongpass",
			true,
			http.StatusUnauthorized,
		},
		{
			"successful creation of service with no description",
			"test_service_4",
			"test_user",
			"",
			"service_pass_4",
			"aes-gcm",
			testPass,
			true,
			http.StatusCreated,
		},
		{
			"invalid algo",
			"test_service_5",
			"test_user",
			"description",
			"service_pass_5",
			"aes",
			testPass,
			true,
			http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t, testDB, "services")

			serviceRequest := ServiceRequest{
				Service:             tt.service,
				Username:            tt.username,
				Description:         tt.description,
				Password:            tt.password,
				EncryptionAlgorithm: tt.encryptionAlgorithm,
				UserPassword:        tt.userPassword,
			}

			requestJSON, _ := json.Marshal(serviceRequest)
			req := httptest.NewRequest(http.MethodPost, "/api/services", strings.NewReader(string(requestJSON)))
			req.Header.Set("Content-Type", "application/json")
			if tt.authorized {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
			}
			rr := httptest.NewRecorder()

			server.HTTPServer.Handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantCode, rr.Code, "unexpected status code")
		})
	}

	cleanupTestDB(t, testDB, "users")
}
