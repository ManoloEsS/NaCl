package server

import (
	"fmt"
	"strings"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
)

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
