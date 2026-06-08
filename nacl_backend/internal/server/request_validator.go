package server

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
)

// validator interface to validate objects coming from the client
// client - server shared shapes

// takes a Validator interaface and a reader and returns an instance of
// the Validator and an error. Used in handlers to decode and validate request body
// into a preset request struct
func decodeAndValidate[T Validator](body io.Reader) (T, error) {
	var req T

	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return req, err
	}

	if err := req.Validate(); err != nil {
		return req, err
	}

	return req, nil
}

type Validator interface {
	Validate() error
}
type UserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ServiceRequest struct {
	Service             string `json:"service"`
	Username            string `json:"username"`
	Description         string `json:"description,omitempty"`
	Password            string `json:"password"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	UserPassword        string `json:"user_password"`
}

type CredentialsRequest struct {
	Password string `json:"password"`
}

func (ur *UserRequest) Validate() error {
	if strings.TrimSpace(ur.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if ur.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (sr *ServiceRequest) Validate() error {
	if strings.TrimSpace(sr.Service) == "" {
		return fmt.Errorf("service name is required")
	}

	if strings.TrimSpace(sr.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if strings.TrimSpace(sr.Password) == "" {
		return fmt.Errorf("password is required")
	}

	if _, err := encryption.ValidAlgorithm(strings.TrimSpace(sr.EncryptionAlgorithm)); err != nil {
		return fmt.Errorf("invalid encryption algorithm")
	}

	if strings.TrimSpace(sr.UserPassword) == "" {
		return fmt.Errorf("user password is required")
	}

	return nil
}

func (cr *CredentialsRequest) Validate() error {
	if strings.TrimSpace(cr.Password) == "" {
		return fmt.Errorf("user password is required")
	}

	return nil
}
