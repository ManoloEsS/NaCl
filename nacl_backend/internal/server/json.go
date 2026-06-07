package server

import (
	"encoding/json"
	"net/http"
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
	s.Logger.Error(msg, "error", err)
	type errorResponse struct {
		Error string `json:"error"`
	}
	s.RespondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}
