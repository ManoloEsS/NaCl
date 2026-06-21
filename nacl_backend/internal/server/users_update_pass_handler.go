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

func (s *Server) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
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

	newPassData, err := dto.DecodeAndValidate[*dto.UpdatePasswordRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not process payload: %w", err),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(w, http.StatusBadRequest, "could not update user password", err)
		return
	}

	err = s.Svc.UpdateUserPassword(r.Context(), userID, newPassData.UserPassword, newPassData.NewPassword)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			s.RespondWithError(
				w, http.StatusForbidden,
				"could not update password",
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
				w, http.StatusForbidden,
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
			fmt.Errorf("could not update user passsword: %w", err),
			"userID", userID.String(),
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w, http.StatusInternalServerError,
			"could not update user",
			err,
		)
		return
	}

	err = s.Svc.SaveOperation(r.Context(), "change password", "nil", userID, uuid.Nil)
	if err != nil {
		s.Logger.Error("could not save operation", "type", "change password", "error", err)
	}

	s.RespondWithJSON(w, http.StatusOK, nil)

}
