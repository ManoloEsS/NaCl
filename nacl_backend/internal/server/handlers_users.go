package server

import (
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	pkgerr "github.com/pkg/errors"
)

func (s *Server) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	userData, err := decodeAndValidate[*CreateUserRequest](r.Body)
	if err != nil {
		s.RespondWithError(w, http.StatusBadRequest, "'username' and 'password' required", err)
		return
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

	queries := s.Db.Queries()
	created, err := queries.CreateUser(
		r.Context(), db.CreateUserParams{
			Username:     userData.Username,
			PasswordHash: hashedPassword,
		})
	if err != nil {
		s.RespondWithError(w, 500, "could not create user", err)
		return
	}

	s.RespondWithJSON(w, 201, created)
}
