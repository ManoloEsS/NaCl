package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/apperr"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
)

func (s *Server) handlerLogin(w http.ResponseWriter, r *http.Request) {
	endpointData := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	loginData, err := decodeAndValidate[*UserRequest](r.Body)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not validate login data: %w", err),
			"endpoint", endpointData,
		)
		s.RespondWithError(w, http.StatusBadRequest, "could not login", err)
		return
	}

	query := s.Db.Queries()
	user, err := query.GetUserByUsername(r.Context(), loginData.Username)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not retrieve user: %w", err),
			"user", loginData.Username,
			"endpoint", endpointData,
		)
		s.RespondWithError(w, http.StatusUnauthorized, "could not log in", err)
		return
	}

	match, err := auth.CheckPasswordHash(loginData.Password, user.PasswordHash)
	if !match || err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("invalid credentials: %w", err),
			"user", loginData.Username,
			"endpoint", endpointData,
		)
		s.RespondWithError(w, http.StatusUnauthorized, "could not log in", err)
		return
	}

	token, err := auth.MakeJWT(user.ID, s.Config.JwtSecret, time.Minute*30)
	if err != nil {
		err = apperr.WithAttrs(
			fmt.Errorf("could not produce token: %w", err),
			"user", loginData.Username,
			"endpoint", endpointData,
		)
		s.RespondWithError(w, http.StatusInternalServerError, "could not log in", err)
		return
	}

	userResponse := UserResponse{user.ID, user.Username}

	s.RespondWithJSON(w, 200, LoginResponse{userResponse, token})
}
