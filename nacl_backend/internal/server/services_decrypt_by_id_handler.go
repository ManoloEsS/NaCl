package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handlerDecryptById(w http.ResponseWriter, r *http.Request) {
	endpointReqPath := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	userID, ok := auth.UserIDFromContext(r.Context())

	if userID == uuid.Nil || !ok {
		err := apperr.WithAttrs(
			fmt.Errorf("could not get user id: %w", errInvalidUserID),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusUnauthorized,
			"not authorized",
			err,
		)
		return
	}

	decryptReq, err := decodeAndValidate[*CredentialsRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusBadRequest,
			"could not retrieve credentials",
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

	match, err := auth.CheckPasswordHash(decryptReq.Password, userData.PasswordHash)
	if !match || err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("password does not match user password: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusUnauthorized,
			"could not retrieve credentials",
			err,
		)
		return
	}
	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("invalid service id: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusBadRequest,
			"could not retrieve credentials",
			err,
		)
		return
	}

	service, err := query.GetServiceById(r.Context(), serviceID)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not find service: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusNotFound,
			"could not retrieve credentials",
			err,
		)
		return
	}

	if userID != service.UserID {
		err = apperr.WithAttrs(
			fmt.Errorf("user with id %s does not own the service with id %s: %w", userID.String(), service.ID.String(), err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusUnauthorized,
			"could not retrieve credentials",
			err,
		)
		return
	}

	decryptedCredentials, err := newServiceCredentialResponse(decryptReq.Password, service, &userData)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not decrypt credentials: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not retrieve credentials",
			err,
		)
		return
	}

	s.RespondWithJSON(w, http.StatusOK, decryptedCredentials)
}

func newServiceCredentialResponse(
	password string,
	service db.Service,
	user *db.User) (ServiceCredentialsResponse, error) {
	decodedSalt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)
	if err != nil {
		return ServiceCredentialsResponse{}, fmt.Errorf("could not decode master key salt: %w", err)
	}

	derivedKey, err := encryption.DeriveKey(password, decodedSalt)
	if err != nil {
		return ServiceCredentialsResponse{}, fmt.Errorf("could not derive key: %w", err)
	}

	decodedMasterKey, err := base64.StdEncoding.DecodeString(user.EncryptedMasterKey)
	if err != nil {
		return ServiceCredentialsResponse{}, fmt.Errorf("could not decode encrypted master key: %w", err)
	}

	masterKey, err := encryption.Decrypt(decodedMasterKey, derivedKey)
	if err != nil {
		return ServiceCredentialsResponse{}, fmt.Errorf("could not decrypt master key: %w", err)
	}

	decryptedUsername, err := encryption.Decrypt(service.EncryptedServiceUsername, masterKey)
	if err != nil {
		return ServiceCredentialsResponse{}, fmt.Errorf("could not decrypt username: %w", err)
	}

	decryptedPassword, err := encryption.Decrypt(service.EncryptedPassword, masterKey)
	if err != nil {
		return ServiceCredentialsResponse{}, fmt.Errorf("could not decrypt password: %w", err)
	}

	decryptedService := ServiceCredentialsResponse{
		Service:             service.Service,
		ServiceUsername:     string(decryptedUsername),
		Password:            string(decryptedPassword),
		Description:         service.Description.String,
		EncryptionAlgorithm: service.EncryptionAlgorithm,
		CreatedAt:           service.CreatedAt.Time,
		UpdatedAt:           service.UpdatedAt.Time,
	}

	return decryptedService, nil
}
