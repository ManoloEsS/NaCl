package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/service"
	"github.com/google/uuid"
)

func (s *Server) HandleCreateCredential(w http.ResponseWriter, r *http.Request) {
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

	serviceData, err := dto.DecodeAndValidate[*dto.CreateCredentialRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusBadRequest,
			"could not create credential",
			err,
		)
		return
	}

	result, err := s.Svc.CreateCredential(r.Context(), userID, serviceData)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			s.RespondWithError(
				w, http.StatusUnauthorized,
				"could not create credential",
				apperr.WithAttrs(
					fmt.Errorf("invalid credentials: %w", err),
					"userID", userID.String(),
					"endpoint", endpointReqPath,
				),
			)
			return
		}
		if errors.Is(err, service.ErrUserNotFound) {
			s.RespondWithError(
				w, http.StatusUnauthorized,
				"not authorized",
				apperr.WithAttrs(
					fmt.Errorf("user not found: %w", err),
					"userID", userID.String(),
					"endpoint", endpointReqPath,
				),
			)
			return
		}
		err = apperr.WithAttrs(
			err,
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not create credential",
			err,
		)
		return
	}

	err = s.Svc.SaveOperation(r.Context(), "create", result.Service, userID, result.ID)
	if err != nil {
		s.Logger.Error("could not save operation", "type", "create", "service", result.Service)
	}
	s.RespondWithJSON(w, http.StatusCreated, result)
}
