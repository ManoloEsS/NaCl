package server

import (
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

	pgID := pgtype.UUID{
		Bytes: userID,
		Valid: true,
	}

	query := s.Db.Queries()
	userData, err := query.GetUserById(r.Context(), pgID)
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

	serviceData, err := decodeAndValidate[*ServiceRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
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

	derivedKey, err := encryption.DeriveKey(serviceData.UserPassword, []byte(userData.MasterKeySalt))
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not get derived key: %w", err),
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

	masterKey, err := encryption.Decrypt([]byte(userData.EncryptedMasterKey), derivedKey)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not decrypt master key: %w", err),
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

	encryptedUsername, err := encryption.Encrypt([]byte(serviceData.Username), masterKey)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not encrypt username: %w", err),
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

	encrytedPassword, err := encryption.Encrypt([]byte(serviceData.Password), masterKey)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not encrypt password: %w", err),
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

	created, err := query.CreateService(r.Context(), db.CreateServiceParams{
		Service:                  serviceData.Service,
		EncryptedServiceUsername: encryptedUsername,
		EncryptedPassword:        encrytedPassword,
		Description:              pgtype.Text{String: serviceData.Description, Valid: true},
		EncryptionAlgorithm:      serviceData.EncryptionAlgorithm,
		UserID:                   userData.ID,
	})
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
