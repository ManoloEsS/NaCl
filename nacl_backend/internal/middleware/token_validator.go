package middleware

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
)

var (
	ErrNoAuthHeader       = errors.New("could not get authorization header")
	ErrNoBearerToken      = errors.New("authorization header does not contain bearer token")
	ErrInvalidBearerToken = errors.New("bearer token is invalid")
)

type errorResponse struct {
	Error string `json:"error"`
}

func TokenValidator(logger *slog.Logger, secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractToken(r)

			if err != nil {
				sendErrorResponse(err, w, logger)
				return
			}

			userId, err := auth.ValidateJWT(token, secret)
			if err != nil {
				sendErrorResponse(err, w, logger)
				return
			}
			ctx := auth.WithUserId(r.Context(), userId)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

func extractToken(r *http.Request) (string, error) {
	authValue := r.Header.Get("Authorization")
	if authValue == "" {
		return "", ErrNoAuthHeader
	}

	if !strings.HasPrefix(authValue, "Bearer ") {
		return "", ErrNoBearerToken
	}

	bearerParts := strings.Split(authValue, " ")
	if len(bearerParts) != 2 {
		return "", ErrInvalidBearerToken
	}

	return bearerParts[1], nil
}

func sendErrorResponse(err error, w http.ResponseWriter, logger *slog.Logger) {
	logger.Error("invalid token", "error", err)
	response, err := json.Marshal(errorResponse{Error: err.Error()})
	if err != nil {
		logger.Error("Error marshalling JSON", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write(response)
}
