package server

import (
	"fmt"
	"strings"
)

type Validator interface {
	Validate() error
}
type UserRequest struct {
	Username string `json:"username"`
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
