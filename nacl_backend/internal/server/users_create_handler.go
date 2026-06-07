package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
)

func (s *Server) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	endpointReqPath := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	userData, err := decodeAndValidate[*UserRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusBadRequest, "could not create user", err)
		return
	}

	salt := encryption.GenerateRandomBytes(32)
	key := encryption.GenerateRandomBytes(32)

	derivedKey, err := encryption.DeriveKey(userData.Password, salt)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not derive key: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}

	encryptedMasterKey, err := encryption.Encrypt(key, derivedKey)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not encrypt master key: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}

	hashedPassword, err := auth.HashPassword(userData.Password)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not hash password: %w", err),
			"user", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}

	query := s.Db.Queries()
	created, err := query.CreateUser(r.Context(), db.CreateUserParams{
		Username:           userData.Username,
		PasswordHash:       hashedPassword,
		MasterKeySalt:      base64.StdEncoding.EncodeToString(salt),
		EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryptedMasterKey),
	})

	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not insert user record: %w", err),
			"user", created.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}

	s.RespondWithJSON(w, 201, UserResponse{ID: created.ID.Bytes, Username: created.Username})
}
