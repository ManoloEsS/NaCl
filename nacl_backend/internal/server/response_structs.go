package server

import "github.com/google/uuid"

type UserResponse struct {
	Id       uuid.UUID
	Username string
}

type LoginResponse struct {
	UserResponse
	Token string
}
