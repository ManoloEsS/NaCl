package dto

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/google/uuid"
)

type Validator interface {
	Validate() error
}

func DecodeAndValidate[T Validator](body io.Reader) (T, error) {
	var req T

	if err := json.NewDecoder(body).Decode(&req); err != nil {
		return req, err
	}

	if err := req.Validate(); err != nil {
		return req, err
	}

	return req, nil
}

// ----------------------------------------------------------------------------------
// Contracts between backend and frontend
// frontend -> backend
// services take these DTO shapes after they are decoded and validated from r.body
// ----------------------------------------------------------------------------------

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateServiceRequest struct {
	Service             string `json:"service"`
	Username            string `json:"username"`
	Description         string `json:"description,omitempty"`
	Password            string `json:"password"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	UserPassword        string `json:"user_password"`
}

type DecryptServiceRequest struct {
	Password string `json:"password"`
}

type UpdateServiceRequest struct {
	Password            string `json:"password"`
	EncryptionAlgorithm string `json:"encryption_algorithm"`
	UserPassword        string `json:"user_password"`
}

func (r *CreateUserRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if r.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (r *CreateServiceRequest) Validate() error {
	if strings.TrimSpace(r.Service) == "" {
		return fmt.Errorf("service name is required")
	}

	if strings.TrimSpace(r.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if strings.TrimSpace(r.Password) == "" {
		return fmt.Errorf("password is required")
	}

	if _, err := encryption.ValidAlgorithm(strings.TrimSpace(r.EncryptionAlgorithm)); err != nil {
		return fmt.Errorf("invalid encryption algorithm")
	}

	if strings.TrimSpace(r.UserPassword) == "" {
		return fmt.Errorf("user password is required")
	}

	return nil
}

func (r *DecryptServiceRequest) Validate() error {
	if strings.TrimSpace(r.Password) == "" {
		return fmt.Errorf("user password is required")
	}

	return nil
}

func (r *UpdateServiceRequest) Validate() error {
	if strings.TrimSpace(r.Password) == "" {
		return fmt.Errorf("password is required")
	}

	if _, err := encryption.ValidAlgorithm(strings.TrimSpace(r.EncryptionAlgorithm)); err != nil {
		return fmt.Errorf("invalid encryption algorithm")
	}

	if strings.TrimSpace(r.UserPassword) == "" {
		return fmt.Errorf("user password is required")
	}

	return nil
}

// -----------------------------------------------------------------------------
// Contracts between backend and frontend
// backend -> frontend
// services return these DTO shapes that are then sent to the frontend directly
// -----------------------------------------------------------------------------

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

type LoginResponse struct {
	UserResponse
	Token string `json:"token"`
}

type ServiceMetadataResponse struct {
	ID                  uuid.UUID `json:"id"`
	Service             string    `json:"service"`
	Description         string    `json:"description"`
	EncryptionAlgorithm string    `json:"encryption_algorithm"`
}

type ServiceCredentialsResponse struct {
	Service             string    `json:"service"`
	ServiceUsername     string    `json:"service_username"`
	Description         string    `json:"description"`
	Password            string    `json:"password"`
	EncryptionAlgorithm string    `json:"encryption_algorithm"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
