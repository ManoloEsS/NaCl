package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/dto"
	servicesvc "github.com/ManoloEsS/NaCl/nacl_backend/internal/service"
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

	userData, err := s.Svc.GetUserData(r.Context(), userID)
	if err != nil {
		err := apperr.WithAttrs(
			fmt.Errorf("could not get user data: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusInternalServerError,
			"could not delete credentials",
			err,
		)
		return
	}

	match, err := auth.CheckPasswordHash(req.UserPassword, userData.PasswordHash)
	if !match || err != nil {
		err := apperr.WithAttrs(
			fmt.Errorf("invalid credentials: %w", err),
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

	credential, err := s.Svc.GetCredentialByID(r.Context(), serviceID)
	if err != nil {
		if errors.Is(err, servicesvc.ErrCredentialNotFound) {
			s.RespondWithJSON(w, http.StatusNoContent, nil)
			return
		}
		err = apperr.WithAttrs(
			fmt.Errorf("could not get credential: %w", err),
			"userID", userID,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusInternalServerError,
			"could not delete credentials", err)
		return
	}

	if credential.UserID != userID {
		s.RespondWithError(w, http.StatusForbidden, "not authorized", nil)
		return
	}

	serviceName, err := s.Svc.DeleteCredentials(r.Context(), serviceID)
	if err != nil {
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

	err = s.Svc.SaveOperation(r.Context(), "delete", serviceName, userID, serviceID)
	if err != nil {
		s.Logger.Error("could not save operation", "type", "delete", "service", serviceName)
	}

	s.RespondWithJSON(w, http.StatusNoContent, nil)

}
