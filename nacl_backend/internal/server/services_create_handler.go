package server

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	errInvalidUserID = errors.New("user id is not valid")
)

func (s *Server) handlerCreateService(w http.ResponseWriter, r *http.Request) {
	endpointReqPath := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	userID, ok := auth.UserIDFromContext(r.Context())

	if userID == uuid.Nil || !ok {
		err := apperr.WithAttrs(
			fmt.Errorf("could not get user id: %w", errInvalidUserID),
			"enpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusUnauthorized,
			"not authorized",
			err,
		)
		return
	}

	query := s.Db.Queries()
	userData, err := query.GetUserById(r.Context(), userID)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not get user data: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not create service",
			err,
		)
		return
	}

	serviceData, err := decodeAndValidate[*NewServiceRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusBadRequest,
			"could not create service",
			err,
		)
		return
	}

	passMatch, err := auth.CheckPasswordHash(serviceData.UserPassword, userData.PasswordHash)
	if !passMatch || err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("password does not match user password: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusUnauthorized,
			"could not create service",
			err,
		)
		return
	}

	newService, err := newCreateServiceParams(serviceData, &userData)
	if err != nil {
		err = apperr.WithAttrs(
			err,
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not create service",
			err,
		)
		return
	}

	created, err := query.CreateService(r.Context(), *newService)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could insert service record: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not create service",
			err,
		)
		return
	}

	s.RespondWithJSON(w, http.StatusCreated, created)
}

func newCreateServiceParams(service *NewServiceRequest, user *db.User) (*db.CreateServiceParams, error) {
	decodedSalt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)
	if err != nil {
		return &db.CreateServiceParams{}, fmt.Errorf("could not decode master key salt: %w", err)
	}

	derivedKey, err := encryption.DeriveKey(service.UserPassword, decodedSalt)
	if err != nil {
		return &db.CreateServiceParams{}, fmt.Errorf("could not derive key: %w", err)
	}

	decodedMasterKey, err := base64.StdEncoding.DecodeString(user.EncryptedMasterKey)
	if err != nil {
		return &db.CreateServiceParams{}, fmt.Errorf("could not decode encrypted master key: %w", err)
	}

	masterKey, err := encryption.Decrypt(decodedMasterKey, derivedKey)
	if err != nil {
		return &db.CreateServiceParams{}, fmt.Errorf("could not decrypt master key: %w", err)
	}

	encryptedUsername, err := encryption.Encrypt([]byte(service.Username), masterKey)
	if err != nil {
		return &db.CreateServiceParams{}, fmt.Errorf("could not encrypt username: %w", err)
	}

	encrytedPassword, err := encryption.Encrypt([]byte(service.Password), masterKey)
	if err != nil {
		return &db.CreateServiceParams{}, fmt.Errorf("could not encrypt password: %w", err)
	}

	newService := db.CreateServiceParams{
		Service:                  service.Service,
		EncryptedServiceUsername: encryptedUsername,
		EncryptedPassword:        encrytedPassword,
		Description:              pgtype.Text{String: service.Description, Valid: true},
		EncryptionAlgorithm:      service.EncryptionAlgorithm,
		UserID:                   user.ID,
	}

	return &newService, nil
}
