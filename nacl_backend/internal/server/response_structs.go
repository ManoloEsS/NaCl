package server

import (
	"github.com/google/uuid"
)

// structs for returned data to client

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
