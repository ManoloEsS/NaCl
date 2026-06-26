package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) HandleUpdateCredentialPassword(w http.ResponseWriter, r *http.Request) {
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

	serviceReq, err := dto.DecodeAndValidate[*dto.UpdateCredentialRequest](r.Body)
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

	credentialID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("invalid credential id: %w", err),
			"userID", userID,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusBadRequest,
			"could not update credential",
			err,
		)
		return
	}

	result, err := s.Svc.UpdateCredentialPassword(r.Context(), userID, credentialID, serviceReq)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			s.RespondWithError(
				w, http.StatusUnauthorized,
				"could not update credential",
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
			fmt.Errorf("could not update credential: %w", err),
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

	err = s.Svc.SaveOperation(r.Context(), service.TypeUpdate, result.Service, userID, credentialID)
	if err != nil {
		s.Logger.Error("could not save operation", "type", service.TypeUpdate.String(), "service", result.Service)
	}

	s.RespondWithJSON(w, http.StatusOK, result)
}
