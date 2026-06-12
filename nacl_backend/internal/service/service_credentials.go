package service

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func (svc *Service) ListServices(ctx context.Context, userID uuid.UUID) ([]dto.ServiceMetadataResponse, error) {
	services, err := svc.Db.Queries().GetAllServicesForUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ServiceMetadataResponse, len(services))
	for i, s := range services {
		var description string
		if s.Description.Valid {
			description = s.Description.String
		}

		result[i] = dto.ServiceMetadataResponse{
			ID:                  s.ID,
			Service:             s.Service,
			Description:         description,
			EncryptionAlgorithm: s.EncryptionAlgorithm,
		}
	}

	return result, nil
}

func (svc *Service) CreateService(ctx context.Context, userID uuid.UUID, req *dto.CreateServiceRequest) (dto.ServiceMetadataResponse, error) {
	user, err := svc.Db.Queries().GetUserById(ctx, userID)
	if err != nil {
		return dto.ServiceMetadataResponse{}, ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(req.UserPassword, user.PasswordHash)
	if err != nil || !match {
		return dto.ServiceMetadataResponse{}, ErrInvalidCredentials
	}

	masterKey, err := decryptMasterKey(req.UserPassword, &user)
	if err != nil {
		return dto.ServiceMetadataResponse{}, err
	}

	encryptedUsername, err := encryption.Encrypt([]byte(req.Username), masterKey)
	if err != nil {
		return dto.ServiceMetadataResponse{}, fmt.Errorf("could not encrypt username: %w", err)
	}

	encryptedPassword, err := encryption.Encrypt([]byte(req.Password), masterKey)
	if err != nil {
		return dto.ServiceMetadataResponse{}, fmt.Errorf("could not encrypt password: %w", err)
	}

	created, err := svc.Db.Queries().CreateService(ctx, db.CreateServiceParams{
		Service:                  req.Service,
		EncryptedServiceUsername: encryptedUsername,
		EncryptedPassword:        encryptedPassword,
		Description:              pgtype.Text{String: req.Description, Valid: true},
		EncryptionAlgorithm:      req.EncryptionAlgorithm,
		UserID:                   user.ID,
	})
	if err != nil {
		return dto.ServiceMetadataResponse{}, fmt.Errorf("could not create service: %w", err)
	}

	var description string
	if created.Description.Valid {
		description = created.Description.String
	}

	return dto.ServiceMetadataResponse{
		ID:                  created.ID,
		Service:             created.Service,
		Description:         description,
		EncryptionAlgorithm: created.EncryptionAlgorithm,
	}, nil
}

func decryptMasterKey(password string, user *db.User) ([]byte, error) {
	decodedSalt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)
	if err != nil {
		return nil, fmt.Errorf("could not decode master key salt: %w", err)
	}

	derivedKey, err := encryption.DeriveKey(password, decodedSalt)
	if err != nil {
		return nil, fmt.Errorf("could not derive key: %w", err)
	}

	decodedMasterKey, err := base64.StdEncoding.DecodeString(user.EncryptedMasterKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode encrypted master key: %w", err)
	}

	masterKey, err := encryption.Decrypt(decodedMasterKey, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("could not decrypt master key: %w", err)
	}

	return masterKey, nil
}

func (svc *Service) DecryptServiceByID(ctx context.Context, userID, serviceID uuid.UUID, password string) (dto.ServiceCredentialsResponse, error) {
	user, err := svc.Db.Queries().GetUserById(ctx, userID)
	if err != nil {
		return dto.ServiceCredentialsResponse{}, ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(password, user.PasswordHash)
	if err != nil || !match {
		return dto.ServiceCredentialsResponse{}, ErrInvalidCredentials
	}

	service, err := svc.Db.Queries().GetServiceById(ctx, db.GetServiceByIdParams{ID: serviceID, UserID: userID})
	if err != nil {
		return dto.ServiceCredentialsResponse{}, ErrServiceNotFound
	}

	masterKey, err := decryptMasterKey(password, &user)
	if err != nil {
		return dto.ServiceCredentialsResponse{}, err
	}

	decryptedUsername, err := encryption.Decrypt(service.EncryptedServiceUsername, masterKey)
	if err != nil {
		return dto.ServiceCredentialsResponse{}, fmt.Errorf("could not decrypt username: %w", err)
	}

	decryptedPassword, err := encryption.Decrypt(service.EncryptedPassword, masterKey)
	if err != nil {
		return dto.ServiceCredentialsResponse{}, fmt.Errorf("could not decrypt password: %w", err)
	}

	var description string
	if service.Description.Valid {
		description = service.Description.String
	}

	return dto.ServiceCredentialsResponse{
		Service:             service.Service,
		ServiceUsername:     string(decryptedUsername),
		Password:            string(decryptedPassword),
		Description:         description,
		EncryptionAlgorithm: service.EncryptionAlgorithm,
		CreatedAt:           service.CreatedAt.Time,
		UpdatedAt:           service.UpdatedAt.Time,
	}, nil
}

func (svc *Service) UpdateServicePassword(ctx context.Context, userID, serviceID uuid.UUID, req *dto.UpdateServiceRequest) (dto.ServiceMetadataResponse, error) {
	user, err := svc.Db.Queries().GetUserById(ctx, userID)
	if err != nil {
		return dto.ServiceMetadataResponse{}, ErrUserNotFound
	}

	match, err := auth.CheckPasswordHash(req.UserPassword, user.PasswordHash)
	if err != nil || !match {
		return dto.ServiceMetadataResponse{}, ErrInvalidCredentials
	}

	masterKey, err := decryptMasterKey(req.UserPassword, &user)
	if err != nil {
		return dto.ServiceMetadataResponse{}, err
	}

	encryptedNewPass, err := encryption.Encrypt([]byte(req.Password), masterKey)
	if err != nil {
		return dto.ServiceMetadataResponse{}, fmt.Errorf("could not encrypt new password: %w", err)
	}

	updated, err := svc.Db.Queries().UpdateService(ctx, db.UpdateServiceParams{
		EncryptedPassword:   encryptedNewPass,
		EncryptionAlgorithm: req.EncryptionAlgorithm,
		ID:                  serviceID,
		UserID:              user.ID,
	})
	if err != nil {
		return dto.ServiceMetadataResponse{}, fmt.Errorf("could not update service: %w", err)
	}

	var description string
	if updated.Description.Valid {
		description = updated.Description.String
	}

	return dto.ServiceMetadataResponse{
		ID:                  updated.ID,
		Service:             updated.Service,
		Description:         description,
		EncryptionAlgorithm: updated.EncryptionAlgorithm,
	}, nil
}
