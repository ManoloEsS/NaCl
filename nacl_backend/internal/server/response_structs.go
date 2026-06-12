package server

import (
	"time"

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

type ServiceCredentialsResponse struct {
	Service             string    `json:"service"`
	ServiceUsername     string    `json:"service_username"`
	Description         string    `json:"description"`
	Password            string    `json:"password"`
	EncryptionAlgorithm string    `json:"encryption_algorithm"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}
