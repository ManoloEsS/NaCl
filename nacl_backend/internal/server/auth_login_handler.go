package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/service"
)

func (s *Server) HandleLogin(w http.ResponseWriter, r *http.Request) {
	endpointData := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	loginData, err := dto.DecodeAndValidate[*dto.LoginRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not validate login data: %w", err),
			"endpoint", endpointData,
		)
		s.RespondWithError(w, http.StatusBadRequest, "could not login", err)
		return
	}

	result, err := s.Svc.Login(r.Context(), loginData.Username, loginData.UserPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			s.RespondWithError(w, http.StatusUnauthorized, "could not log in", apperr.WithAttrs(
				fmt.Errorf("invalid credentials: %w", err),
				"username", loginData.Username,
				"endpoint", endpointData,
			))
			return
		}
		err = apperr.WithAttrs(
			fmt.Errorf("could not log in: %w", err),
			"username", loginData.Username,
			"endpoint", endpointData,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not log in", err)
		return
	}

	err = s.Svc.SaveOperation(r.Context(), service.TypeLogin, "nil", result.ID)
	if err != nil {
		s.Logger.Error("could not save operation", "type", service.TypeLogin.String(), "error", err)
	}

	s.RespondWithJSON(w, http.StatusOK, result)
}
