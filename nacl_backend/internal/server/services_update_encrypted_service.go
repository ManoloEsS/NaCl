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

func (s *Server) handlerUpdateServicePass(w http.ResponseWriter, r *http.Request) {
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

	serviceReq, err := decodeAndValidate[*UpdateServiceRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusBadRequest,
			"could not update service",
			err,
		)
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("invalid service id: %w", err),
			"userID", userID,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusBadRequest,
			"could not update service",
			err,
		)
		return
	}

	query := s.Db.Queries()
	userData, err := query.GetUserById(r.Context(), userID)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not get user data: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusInternalServerError,
			"could not update service",
			err,
		)
		return
	}

	match, err := auth.CheckPasswordHash(serviceReq.UserPassword, userData.PasswordHash)
	if !match || err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("user password does not match: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusUnauthorized,
			"could not update service",
			err,
		)
		return
	}

	updateServiceParams, err := newUpdateServiceParams(
		serviceID,
		serviceReq,
		&userData,
	)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not create update service parameters struct: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusInternalServerError,
			"could not update service",
			err,
		)
		return
	}

	updatedService, err := query.UpdateService(r.Context(), *updateServiceParams)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not update service in db: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusInternalServerError,
			"could not update service",
			err,
		)
		return
	}

	s.RespondWithJSON(w, http.StatusOK, ServiceMetadataResponse{
		ID:                  updatedService.ID,
		Service:             updatedService.Service,
		Description:         updatedService.Description.String,
		EncryptionAlgorithm: updatedService.EncryptionAlgorithm,
	})

}

func newUpdateServiceParams(
	serviceID uuid.UUID,
	serviceReq *UpdateServiceRequest,
	user *db.User,
) (*db.UpdateServiceParams, error) {
	decodedSalt, err := base64.StdEncoding.DecodeString(user.MasterKeySalt)
	if err != nil {
		return &db.UpdateServiceParams{}, fmt.Errorf("could not decode master key salt: %w", err)
	}

	derivedKey, err := encryption.DeriveKey(serviceReq.UserPassword, decodedSalt)
	if err != nil {
		return &db.UpdateServiceParams{}, fmt.Errorf("could not derive key: %w", err)
	}

	decodedMasterKey, err := base64.StdEncoding.DecodeString(user.EncryptedMasterKey)
	if err != nil {
		return &db.UpdateServiceParams{}, fmt.Errorf("could not decode encrypted master key: %w", err)
	}

	masterKey, err := encryption.Decrypt(decodedMasterKey, derivedKey)
	if err != nil {
		return &db.UpdateServiceParams{}, fmt.Errorf("could not decrypt master key: %w", err)
	}

	encryptedNewPass, err := encryption.Encrypt([]byte(serviceReq.Password), masterKey)
	if err != nil {
		return &db.UpdateServiceParams{}, fmt.Errorf("could not encrypt new password: %w", err)
	}

	updateService := db.UpdateServiceParams{
		EncryptedPassword:   encryptedNewPass,
		EncryptionAlgorithm: serviceReq.EncryptionAlgorithm,
		ID:                  serviceID,
		UserID:              user.ID,
	}
	return &updateService, nil
}
