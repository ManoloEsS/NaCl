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

func (s *Server) HandleDeleteCredential(w http.ResponseWriter, r *http.Request) {
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

	req, err := dto.DecodeAndValidate[*dto.DeleteCredentialsRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusBadRequest,
			"could not retrieve credentials",
			err,
		)
		return
	}

	serviceID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("invalid service id: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusBadRequest,
			"could not delete credentials",
			err,
		)
		return
	}

	serviceName, err := s.Svc.DeleteCredentials(r.Context(), userID, serviceID, req.UserPassword)
	if err != nil {
		if errors.Is(err, service.ErrCredentialNotFound) {
			s.RespondWithJSON(w, http.StatusNoContent, nil)
			return
		}

		if errors.Is(err, service.ErrUnauthorized) {
			err := apperr.WithAttrs(
				fmt.Errorf("user not authorized: %w", err),
				"userID", userID.String(),
				"endpoint", endpointReqPath,
			)
			s.RespondWithError(
				w, http.StatusForbidden,
				"could not delete credentials",
				err,
			)
			return
		}

		err = apperr.WithAttrs(
			fmt.Errorf("could not delete credentials from db: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not delete credentials",
			err,
		)
		return
	}

	err = s.Svc.SaveOperation(r.Context(), service.TypeDelete, serviceName, userID, serviceID)
	if err != nil {
		s.Logger.Error("could not save operation", "type", service.TypeDelete.String(), "service", serviceName)
	}

	s.RespondWithJSON(w, http.StatusNoContent, nil)

}
