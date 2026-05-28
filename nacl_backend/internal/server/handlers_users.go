package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	pkgerr "github.com/pkg/errors"
)

func (s *Server) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var userRequest CreateUserRequest
	err := decoder.Decode(&userRequest)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "'username' and 'password' required", err)
		return
	}

	// Validate username
	if strings.TrimSpace(userRequest.Username) == "" {
		s.RespondWithError(w, http.StatusBadRequest, "username is required", nil)
		return
	}

	// Validate password
	if userRequest.Password == "" {
		s.RespondWithError(w, http.StatusBadRequest, "password is required", nil)
		return
	}

	hashedPassword, err := s.hashPassword(userRequest.Password)
	if err != nil {
		s.RespondWithError(w, http.StatusInternalServerError, "could not process password", pkgerr.WithStack(err))
		return
	}

	queries := s.Db.Queries()
	created, err := queries.CreateUser(r.Context(), db.CreateUserParams{Username: userRequest.Username, PasswordHash: hashedPassword})
	if err != nil {
		s.RespondWithError(w, 500, "could not create user", pkgerr.WithStack(err))
		return
	}

	s.RespondWithJSON(w, 201, created)
}
