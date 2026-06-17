package server

import (
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
)

func (s *Server) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	endpointReqPath := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	userData, err := dto.DecodeAndValidate[*dto.CreateUserRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusBadRequest, "could not create user", err)
		return
	}

	err = s.Svc.CreateUser(r.Context(), userData.Username, userData.UserPassword)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not create user: %w", err),
			"username", userData.Username,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not create user", err)
		return
	}

	s.RespondWithJSON(w, http.StatusCreated, nil)
}
