package server

import (
	"encoding/base64"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	pkgerr "github.com/pkg/errors"
)

func (s *Server) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	userData, err := decodeAndValidate[*CreateUserRequest](r.Body)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "'username' and 'password' required", err)
		return
	}

	salt := encryption.GenerateRandomBytes(32)
	key := encryption.GenerateRandomBytes(32)

	derivedKey, err := encryption.DeriveKey(userData.Password, salt)
	if err != nil {
		s.RespondWithError(w, 500, "couldn't derive key", err)
		s.Logger.Error("could not derive key", "error", err)
	}

	encryptedMasterKey, err := encryption.Encrypt(key, derivedKey)
	if err != nil {
		s.RespondWithError(w, 500, "could not encrypt master key", err)
		s.Logger.Error("could not encrypt master key", "error", err)
	}

	hashedPassword, err := s.hashPassword(userData.Password)
	if err != nil {
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not process password",
			pkgerr.WithStack(err),
		)
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
		s.RespondWithError(w, 500, "could not create user", err)
		return
	}

	s.RespondWithJSON(w, 201, created)
}
