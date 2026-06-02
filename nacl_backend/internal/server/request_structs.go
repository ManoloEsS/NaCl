package server

import (
	"fmt"
	"strings"
)

type Validator interface {
	Validate() error
}
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (ur *CreateUserRequest) Validate() error {
	if strings.TrimSpace(ur.Username) == "" {
		return fmt.Errorf("username is required")
	}

	if ur.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}
