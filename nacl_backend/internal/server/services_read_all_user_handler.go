package server

import (
	"fmt"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/google/uuid"
)

func (s *Server) handlerGetAllServicesUser(w http.ResponseWriter, r *http.Request) {
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

	query := s.Db.Queries()
	services, err := query.GetAllServicesForUserId(r.Context(), userID)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not get services for user: %w", err),
			"userID", userID,
			"endpoint", endpointReqPath,
		)
		s.RespondWithError(
			w,
			http.StatusInternalServerError,
			"could not retrieve user data",
			err,
		)
		return
	}

	processedServices := make([]ServiceMetadataResponse, len(services))
	for i, service := range services {
		var description string

		if service.Description.Valid {
			description = service.Description.String
		}

		processed := ServiceMetadataResponse{
			ID:                  service.ID,
			Service:             service.Service,
			Description:         description,
			EncryptionAlgorithm: service.EncryptionAlgorithm,
		}

		processedServices[i] = processed
	}

	s.RespondWithJSON(w, 200, processedServices)
}
