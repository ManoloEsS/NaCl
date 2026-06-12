package service

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
)

func (svc *Service) CreateUser(ctx context.Context, username, password string) (dto.UserResponse, error) {
	salt, err := encryption.GenerateRandomBytes(32)
	if err != nil {
		return dto.UserResponse{}, err
	}

	key, err := encryption.GenerateRandomBytes(32)
	if err != nil {
		return dto.UserResponse{}, err
	}

	derivedKey, err := encryption.DeriveKey(password, salt)
	if err != nil {
		return dto.UserResponse{}, err
	}

	encryptedMasterKey, err := encryption.Encrypt(key, derivedKey)
	if err != nil {
		return dto.UserResponse{}, err
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return dto.UserResponse{}, err
	}

	created, err := svc.Db.Queries().CreateUser(ctx, db.CreateUserParams{
		Username:           username,
		PasswordHash:       hashedPassword,
		MasterKeySalt:      base64.StdEncoding.EncodeToString(salt),
		EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryptedMasterKey),
	})
	if err != nil {
		return dto.UserResponse{}, err
	}

	return dto.UserResponse{ID: created.ID, Username: created.Username}, nil
}

func (svc *Service) Login(ctx context.Context, username, password string) (dto.LoginResponse, error) {
	user, err := svc.Db.Queries().GetUserByUsername(ctx, username)
	if err != nil {
		return dto.LoginResponse{}, ErrInvalidCredentials
	}

	match, err := auth.CheckPasswordHash(password, user.PasswordHash)
	if err != nil || !match {
		return dto.LoginResponse{}, ErrInvalidCredentials
	}

	token, err := auth.MakeJWT(user.ID, svc.Config.JwtSecret, time.Minute*30)
	if err != nil {
		return dto.LoginResponse{}, err
	}

	return dto.LoginResponse{
		UserResponse: dto.UserResponse{ID: user.ID, Username: user.Username},
		Token:        token,
	}, nil
}
