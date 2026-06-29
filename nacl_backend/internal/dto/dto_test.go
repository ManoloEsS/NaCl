package dto

import "testing"

func TestDecodeAndValidate(t *testing.T) {
	// TODO: valid CreateUserRequest → no error
	// TODO: invalid JSON → error
	// TODO: empty body → error
	// TODO: valid JSON but Validate fails (empty username) → error
}

func TestCreateUserRequest_Validate(t *testing.T) {
	// TODO: valid
	// TODO: empty username
	// TODO: empty password
}

func TestLoginRequest_Validate(t *testing.T) {
	// TODO: valid
	// TODO: empty username
	// TODO: empty password
}

func TestUpdatePasswordRequest_Validate(t *testing.T) {
	// TODO: valid
	// TODO: empty user_password
	// TODO: empty new_password
}

func TestCreateCredentialRequest_Validate(t *testing.T) {
	// TODO: valid (algo: "aes-gcm")
	// TODO: empty service
	// TODO: empty service_username
	// TODO: empty service_password
	// TODO: invalid algorithm
	// TODO: empty user_password
}

func TestDecryptCredentialRequest_Validate(t *testing.T) {
	// TODO: valid
	// TODO: empty user_password
}

func TestUpdateCredentialRequest_Validate(t *testing.T) {
	// TODO: valid (algo: "aes-gcm")
	// TODO: empty password
	// TODO: invalid algorithm
	// TODO: empty user_password
}
