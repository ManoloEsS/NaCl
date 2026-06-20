package service

import (
	"context"
	"encoding/base64"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/google/uuid"
)

func (svc *Service) CreateUser(ctx context.Context, username, password string) error {
	salt, err := encryption.GenerateRandomBytes(32)
	if err != nil {
		return err
	}

	key, err := encryption.GenerateRandomBytes(32)
	if err != nil {
		return err
	}

	derivedKey, err := encryption.DeriveKey(password, salt)
	if err != nil {
		return err
	}

	encryptedMasterKey, err := encryption.Encrypt(key, derivedKey)
	if err != nil {
		return err
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return err
	}

	err = svc.Db.Queries().CreateUser(ctx, db.CreateUserParams{
		Username:           username,
		PasswordHash:       hashedPassword,
		MasterKeySalt:      base64.StdEncoding.EncodeToString(salt),
		EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryptedMasterKey),
	})
	if err != nil {
		return err
	}

	return nil
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

func (svc *Service) UpdateUserPassword(ctx context.Context, userID uuid.UUID, oldPass, newPass string) error {
	user, err := svc.Db.Queries().GetUserById(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(oldPass, user.PasswordHash)
	if err != nil || !match {
		return ErrInvalidCredentials
	}

	masterKey, err := decryptMasterKey(oldPass, &user)
	if err != nil {
		return err
	}

	decodedSalt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)
	newDerivedKey, err := encryption.DeriveKey(newPass, decodedSalt)
	if err != nil {
		return err
	}

	encryptedMasterKey, err := encryption.Encrypt(masterKey, newDerivedKey)
	if err != nil {
		return err
	}

	encodedEncryptedKey := base64.StdEncoding.EncodeToString(encryptedMasterKey)

	hashedPassword, err := auth.HashPassword(newPass)
	if err != nil {
		return err
	}

	var newData = db.UpdateUserPassHashAndKeyParams{
		PasswordHash:       hashedPassword,
		EncryptedMasterKey: encodedEncryptedKey,
		ID:                 userID,
	}

	err = svc.Db.Queries().UpdateUserPassHashAndKey(ctx, newData)
	if err != nil {
		return err
	}

	return nil
}
