package server

import (
	"encoding/json"
	"net/http"

	"github.com/alexedwards/argon2id"
)

// RespondWithJSON marshals the payload to JSON and writes it to the response with the given status code.
// Sets Content-Type to application/json and logs errors if marshaling fails.
func (s *Server) RespondWithJSON(w http.ResponseWriter, code int, payload any) {
	response, err := json.Marshal(payload)
	if err != nil {
		s.Logger.Error("Error marshalling JSON", "error", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

// RespondWithError responds with a JSON error message and logs the error if provided.
func (s *Server) RespondWithError(w http.ResponseWriter, code int, msg string, err error) {
	type errorResponse struct {
		Error string `json:"error"`
	}
	s.RespondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

// HashPassword generates a secure Argon2id hash of the provided password.
// Used by HandlerCreateUser and HandlerUserUpdate to store passwords securely.
func (s *Server) hashPassword(password string) (string, error) {
	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hash, nil
}

// CheckPasswordHash compares a plaintext password with an Argon2id hash.
// Used by HandlerUserLogin to verify user credentials during authentication.
// func (s *Server) checkPasswordHash(password, hash string) (bool, error) {
// 	match, err := argon2id.ComparePasswordAndHash(password, hash)
// 	if err != nil {
// 		return match, err
// 	}
//
// 	return match, nil
// }
