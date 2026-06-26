package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (svc *Service) ListCredentials(ctx context.Context, userID uuid.UUID) ([]dto.CredentialMetadataResponse, error) {
	credentials, err := svc.Queries.GetAllCredentialsForUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.CredentialMetadataResponse, len(credentials))
	for i, s := range credentials {
		var description string
		if s.Description.Valid {
			description = s.Description.String
		}

		result[i] = dto.CredentialMetadataResponse{
			ID:                  s.ID,
			Service:             s.Service,
			Description:         description,
			EncryptionAlgorithm: s.EncryptionAlgorithm,
		}
	}

	return result, nil
}

func (svc *Service) CreateCredential(ctx context.Context, userID uuid.UUID, req *dto.CreateCredentialRequest) (dto.CredentialMetadataResponse, error) {
	user, err := svc.Queries.GetUserById(ctx, userID)
	if err != nil {
		return dto.CredentialMetadataResponse{}, ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(req.UserPassword, user.PasswordHash)
	if err != nil || !match {
		return dto.CredentialMetadataResponse{}, ErrInvalidCredentials
	}

	masterKey, err := decryptMasterKey(req.UserPassword, user.MasterKeySalt, user.EncryptedMasterKey)
	if err != nil {
		return dto.CredentialMetadataResponse{}, err
	}

	encryptedUsername, err := encryption.Encrypt([]byte(req.ServiceUsername), masterKey)
	if err != nil {
		return dto.CredentialMetadataResponse{}, fmt.Errorf("could not encrypt username: %w", err)
	}

	encryptedPassword, err := encryption.Encrypt([]byte(req.ServicePassword), masterKey)
	if err != nil {
		return dto.CredentialMetadataResponse{}, fmt.Errorf("could not encrypt password: %w", err)
	}

	created, err := svc.Queries.CreateCredential(ctx, db.CreateCredentialParams{
		Service:                  req.Service,
		EncryptedServiceUsername: encryptedUsername,
		EncryptedPassword:        encryptedPassword,
		Description:              pgtype.Text{String: req.Description, Valid: true},
		EncryptionAlgorithm:      req.EncryptionAlgorithm,
		UserID:                   user.ID,
	})
	if err != nil {
		return dto.CredentialMetadataResponse{}, fmt.Errorf("could not create credential: %w", err)
	}

	var description string
	if created.Description.Valid {
		description = created.Description.String
	}

	return dto.CredentialMetadataResponse{
		ID:                  created.ID,
		Service:             created.Service,
		Description:         description,
		EncryptionAlgorithm: created.EncryptionAlgorithm,
	}, nil
}

func decryptMasterKey(password string, masterKeySalt string, encryptedKey string) ([]byte, error) {
	decodedSalt, err := base64.StdEncoding.DecodeString(masterKeySalt)
	if err != nil {
		return nil, fmt.Errorf("could not decode master key salt: %w", err)
	}

	derivedKey, err := encryption.DeriveKey(password, decodedSalt)
	if err != nil {
		return nil, fmt.Errorf("could not derive key: %w", err)
	}

	decodedMasterKey, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode encrypted master key: %w", err)
	}

	masterKey, err := encryption.Decrypt(decodedMasterKey, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt master key: %w", err)
	}

	return masterKey, nil
}

func (svc *Service) DecryptCredentialByID(ctx context.Context, userID, credentialID uuid.UUID, password string) (dto.DecryptedCredentialResponse, error) {
	user, err := svc.Queries.GetUserById(ctx, userID)
	if err != nil {
		return dto.DecryptedCredentialResponse{}, ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(password, user.PasswordHash)
	if err != nil || !match {
		return dto.DecryptedCredentialResponse{}, ErrInvalidCredentials
	}

	credential, err := svc.Queries.GetCredentialById(ctx, credentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dto.DecryptedCredentialResponse{}, ErrCredentialNotFound
		}
		return dto.DecryptedCredentialResponse{}, err
	}

	masterKey, err := decryptMasterKey(password, user.MasterKeySalt, user.EncryptedMasterKey)
	if err != nil {
		return dto.DecryptedCredentialResponse{}, err
	}

	decryptedUsername, err := encryption.Decrypt(credential.EncryptedServiceUsername, masterKey)
	if err != nil {
		return dto.DecryptedCredentialResponse{}, fmt.Errorf("could not decrypt username: %w", err)
	}

	decryptedPassword, err := encryption.Decrypt(credential.EncryptedPassword, masterKey)
	if err != nil {
		return dto.DecryptedCredentialResponse{}, fmt.Errorf("could not decrypt password: %w", err)
	}

	var description string
	if credential.Description.Valid {
		description = credential.Description.String
	}

	return dto.DecryptedCredentialResponse{
		Service:             credential.Service,
		ServiceUsername:     string(decryptedUsername),
		ServicePassword:     string(decryptedPassword),
		Description:         description,
		EncryptionAlgorithm: credential.EncryptionAlgorithm,
		CreatedAt:           credential.CreatedAt.Time,
		UpdatedAt:           credential.UpdatedAt.Time,
	}, nil
}

func (svc *Service) UpdateCredentialPassword(ctx context.Context, userID, credentialID uuid.UUID, req *dto.UpdateCredentialRequest) (dto.CredentialMetadataResponse, error) {
	user, err := svc.Queries.GetUserById(ctx, userID)
	if err != nil {
		return dto.CredentialMetadataResponse{}, ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(req.UserPassword, user.PasswordHash)
	if err != nil || !match {
		return dto.CredentialMetadataResponse{}, ErrInvalidCredentials
	}

	masterKey, err := decryptMasterKey(req.UserPassword, user.MasterKeySalt, user.EncryptedMasterKey)
	if err != nil {
		return dto.CredentialMetadataResponse{}, err
	}

	encryptedNewPass, err := encryption.Encrypt([]byte(req.ServicePassword), masterKey)
	if err != nil {
		return dto.CredentialMetadataResponse{}, fmt.Errorf("could not encrypt new password: %w", err)
	}

	updated, err := svc.Queries.UpdateCredential(ctx, db.UpdateCredentialParams{
		EncryptedPassword:   encryptedNewPass,
		EncryptionAlgorithm: req.EncryptionAlgorithm,
		ID:                  credentialID,
		UserID:              user.ID,
	})
	if err != nil {
		return dto.CredentialMetadataResponse{}, fmt.Errorf("could not update credential: %w", err)
	}

	var description string
	if updated.Description.Valid {
		description = updated.Description.String
	}

	return dto.CredentialMetadataResponse{
		ID:                  updated.ID,
		Service:             updated.Service,
		Description:         description,
		EncryptionAlgorithm: updated.EncryptionAlgorithm,
	}, nil
}

func (svc *Service) DeleteCredentials(ctx context.Context, credentialID, userID uuid.UUID, password string) (string, error) {
	match, err := svc.VerifyUserPassword(ctx, userID, password)
	if !match || err != nil {
		return "", ErrUnauthorized
	}

	credential, err := svc.Queries.GetCredentialById(ctx, credentialID)
	if err != nil {
		return "", ErrCredentialNotFound
	}

	if userID != credential.UserID {
		return "", ErrUnauthorized
	}

	service, err := svc.Queries.DeleteCredentialById(ctx, credentialID)
	if err != nil {
		return "", err
	}

	return service, nil
}
