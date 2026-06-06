package server

import "github.com/google/uuid"

type UserResponse struct {
	ID       uuid.UUID
	Username string
}

type LoginResponse struct {
	UserResponse
	Token string
}
