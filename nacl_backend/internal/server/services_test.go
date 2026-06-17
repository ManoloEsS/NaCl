package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
)

func TestHandleCreateService(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()
	cleanupTestDB(t, testDB, "users")

	server := newTestServer(t, testDB)

	testUser := "test_services_user"
	testPass := "test_services_pass"

	body := fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "user creation failed")

	body = fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "login failed")

	var loginResp dto.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&loginResp)
	assert.NoError(t, err, "could not decode login response")
	token := loginResp.Token

	tests := []struct {
		name                string
		service             string
		serviceUsername     string
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

			serviceRequest := dto.CreateServiceRequest{
				Service:             tt.service,
				ServiceUsername:     tt.serviceUsername,
				Description:         tt.description,
				ServicePassword:     tt.password,
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

}

func TestHandleListServices(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()
	cleanupTestDB(t, testDB, "users", "services")

	server := newTestServer(t, testDB)

	testUser := "test_services_user"
	testPass := "test_services_pass"

	// create user
	body := fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "user creation failed")

	// login as user
	body = fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "login failed")

	var loginResp dto.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&loginResp)
	assert.NoError(t, err, "could not decode login response")
	token := loginResp.Token

	services := []struct {
		service         string
		serviceUsername string
		description     string
		password        string
		algo            string
		userPassword    string
	}{

		{
			"test_service_1",
			"test_user",
			"description",
			"service_pass_1",
			"aes-gcm",
			testPass,
		},
		{
			"test_service_3",
			"test_user",
			"description",
			"service_pass_3",
			"aes-gcm",
			testPass,
		},
		{
			"test_service_4",
			"test_user",
			"",
			"service_pass_4",
			"aes-gcm",
			testPass,
		},
		{
			"test_service_5",
			"test_user",
			"description",
			"service_pass_5",
			"aes-gcm",
			testPass,
		},
	}

	for _, service := range services {
		serviceRequest := dto.CreateServiceRequest{
			Service:             service.service,
			ServiceUsername:     service.serviceUsername,
			Description:         service.description,
			ServicePassword:     service.password,
			EncryptionAlgorithm: service.algo,
			UserPassword:        service.userPassword,
		}

		requestJSON, _ := json.Marshal(serviceRequest)
		req := httptest.NewRequest(http.MethodPost, "/api/services", strings.NewReader(string(requestJSON)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		rr := httptest.NewRecorder()

		server.HTTPServer.Handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code, "error crating service")
	}

	tests := []struct {
		name      string
		token     string
		wantCode  int
		wantCount int
		setupFunc func(string) string
	}{
		{
			name:      "successfully gets all services for user",
			token:     token,
			wantCode:  http.StatusOK,
			wantCount: len(services),
			setupFunc: returnSameToken,
		},
		{
			name:      "unauthorized - no token",
			token:     "",
			wantCode:  http.StatusUnauthorized,
			wantCount: 0,
			setupFunc: returnSameToken,
		},
		{
			name:      "unauthorized - invalid token",
			token:     "invalid-token",
			wantCode:  http.StatusUnauthorized,
			wantCount: 0,
			setupFunc: returnSameToken,
		},
		{
			name:      "returns empty array when user has no services",
			token:     token,
			wantCode:  http.StatusOK,
			wantCount: 0,
			setupFunc: func(token string) string {
				// create user
				body := fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, "test_2", "test_pass2")
				req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				server.HTTPServer.Handler.ServeHTTP(rr, req)
				assert.Equal(t, http.StatusCreated, rr.Code, "user creation failed")

				// login as user
				body = fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, "test_2", "test_pass2")
				req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
				req.Header.Set("Content-Type", "application/json")
				rr = httptest.NewRecorder()

				server.HTTPServer.Handler.ServeHTTP(rr, req)
				assert.Equal(t, http.StatusOK, rr.Code, "login failed")

				var loginResp dto.LoginResponse
				err := json.NewDecoder(rr.Body).Decode(&loginResp)
				assert.NoError(t, err, "could not decode login response")

				return loginResp.Token
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testToken := tt.setupFunc(tt.token)
			req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testToken))
			rr := httptest.NewRecorder()

			server.HTTPServer.Handler.ServeHTTP(rr, req)
			require.Equal(t, tt.wantCode, rr.Code, "unexpected status code")

			if tt.wantCode != 200 {
				return
			}
			bodyJSON, err := io.ReadAll(rr.Body)
			require.NoError(t, err, "error reading recorder body")
			var servicesResponse []dto.ServiceMetadataResponse
			err = json.Unmarshal(bodyJSON, &servicesResponse)
			require.NoError(t, err, "unexpected error")

			assert.Equal(t, tt.wantCount, len(servicesResponse))
			if tt.wantCount > 0 {
				assert.Equal(t, services[0].service, servicesResponse[0].Service)
				assert.Equal(t, services[0].description, servicesResponse[0].Description)
				assert.Equal(t, services[0].algo, servicesResponse[0].EncryptionAlgorithm)
			}
		})
	}

}

func TestHandleDecryptServiceByID(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()
	cleanupTestDB(t, testDB, "users", "services")

	server := newTestServer(t, testDB)

	testUser := "test_services_user"
	testPass := "test_services_pass"

	// create user
	body := fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "user creation failed")

	// login as user
	body = fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "login failed")

	var loginResp dto.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&loginResp)
	assert.NoError(t, err, "could not decode login response")
	token := loginResp.Token

	// create service
	serviceRequest := dto.CreateServiceRequest{
		Service:             "test_service_1",
		ServiceUsername:     "test_user",
		Description:         "description",
		ServicePassword:     "service_pass_1",
		EncryptionAlgorithm: "aes-gcm",
		UserPassword:        testPass,
	}

	requestJSON, _ := json.Marshal(serviceRequest)
	req = httptest.NewRequest(http.MethodPost, "/api/services", strings.NewReader(string(requestJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "error crating service")

	var serviceData dto.ServiceMetadataResponse
	bodyReader, err := io.ReadAll(rr.Body)
	assert.NoError(t, err, "unexpected error")

	err = json.Unmarshal(bodyReader, &serviceData)
	assert.NoError(t, err, "unexpected error")

	tests := []struct {
		name      string
		token     string
		userPass  string
		serviceID uuid.UUID
		wantCode  int
	}{
		{
			name:      "successfully decrypts and responds with decrypted service data",
			token:     token,
			userPass:  testPass,
			serviceID: serviceData.ID,
			wantCode:  200,
		},
		{
			name:      "unauthorized request fails - wrong token",
			token:     "wrong-token",
			userPass:  testPass,
			serviceID: serviceData.ID,
			wantCode:  401,
		},
		{
			name:      "unauthorized request fails - wrong user password in body",
			token:     token,
			userPass:  "wrong_pass",
			serviceID: serviceData.ID,
			wantCode:  401,
		},
		{
			name:      "invalid service id request fails",
			token:     token,
			userPass:  testPass,
			serviceID: uuid.New(),
			wantCode:  404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decryptRequest := dto.DecryptServiceRequest{
				UserPassword: tt.userPass,
			}
			requestJSON, _ := json.Marshal(decryptRequest)
			urlPath := fmt.Sprintf("/api/services/%s/credentials", tt.serviceID.String())
			req = httptest.NewRequest(http.MethodPost, urlPath, strings.NewReader(string(requestJSON)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			rr = httptest.NewRecorder()

			server.HTTPServer.Handler.ServeHTTP(rr, req)
			assert.Equal(t, tt.wantCode, rr.Code)

			if tt.wantCode == 200 {
				bodyJSON, err := io.ReadAll(rr.Body)
				assert.NoError(t, err, "error reading recorder body")
				var credentialsResponse dto.ServiceCredentialsResponse
				err = json.Unmarshal(bodyJSON, &credentialsResponse)
				assert.NoError(t, err, "unexpected error")

				assert.Equal(t, serviceRequest.ServiceUsername, credentialsResponse.ServiceUsername)
				assert.Equal(t, serviceRequest.ServicePassword, credentialsResponse.ServicePassword)
				assert.Equal(t, serviceRequest.EncryptionAlgorithm, credentialsResponse.EncryptionAlgorithm)
			}
		})
	}

}

func TestHandleUpdateServicePassword(t *testing.T) {
	testDB := newTestDB(t)
	defer testDB.Close()
	cleanupTestDB(t, testDB, "users", "services")

	server := newTestServer(t, testDB)

	testUser := "test_services_user"
	testPass := "test_services_pass"

	// create user
	body := fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "user creation failed")

	// login as user
	body = fmt.Sprintf(`{"username": "%s", "user_password": "%s"}`, testUser, testPass)
	req = httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code, "login failed")

	var loginResp dto.LoginResponse
	err := json.NewDecoder(rr.Body).Decode(&loginResp)
	assert.NoError(t, err, "could not decode login response")
	token := loginResp.Token

	// create service
	serviceRequest := dto.CreateServiceRequest{
		Service:             "test_service_1",
		ServiceUsername:     "test_user",
		Description:         "description",
		ServicePassword:     "service_pass_1",
		EncryptionAlgorithm: "aes-gcm",
		UserPassword:        testPass,
	}

	requestJSON, _ := json.Marshal(serviceRequest)
	req = httptest.NewRequest(http.MethodPost, "/api/services", strings.NewReader(string(requestJSON)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	rr = httptest.NewRecorder()

	server.HTTPServer.Handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code, "error crating service")

	var serviceData dto.ServiceMetadataResponse
	bodyReader, err := io.ReadAll(rr.Body)
	assert.NoError(t, err, "unexpected error")

	err = json.Unmarshal(bodyReader, &serviceData)
	assert.NoError(t, err, "unexpected error")

	serviceID := serviceData.ID

	tests := []struct {
		name          string
		token         string
		updateRequest dto.UpdateServiceRequest
		serviceID     uuid.UUID
		wantCode      int
	}{
		{
			name:  "successfully updates password of service",
			token: token,
			updateRequest: dto.UpdateServiceRequest{
				ServicePassword:     "new_password",
				EncryptionAlgorithm: "aes-gcm",
				UserPassword:        testPass,
			},
			serviceID: serviceID,
			wantCode:  200,
		},
		{
			name:  "fails with invalid request object",
			token: token,
			updateRequest: dto.UpdateServiceRequest{
				ServicePassword: "new_pass",
				UserPassword:    testPass,
			},
			serviceID: serviceID,
			wantCode:  400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestJSON, _ := json.Marshal(tt.updateRequest)
			urlPath := fmt.Sprintf("/api/services/%s", tt.serviceID.String())
			req = httptest.NewRequest(http.MethodPatch, urlPath, strings.NewReader(string(requestJSON)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tt.token))
			rr = httptest.NewRecorder()

			if tt.wantCode == 200 {
				server.HTTPServer.Handler.ServeHTTP(rr, req)
				assert.Equal(t, tt.wantCode, rr.Code)

				var updatedServiceData dto.ServiceMetadataResponse
				bodyReader, err := io.ReadAll(rr.Body)
				assert.NoError(t, err, "unexpected error")

				err = json.Unmarshal(bodyReader, &updatedServiceData)
				assert.NoError(t, err, "unexpected error")

				assert.Equal(t, serviceData, updatedServiceData, "should not be equal")
			}
		})
	}

}

func returnSameToken(token string) string {
	return token
}
