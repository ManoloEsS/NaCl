package server

import "net/http"

func (s *Server) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.RespondWithJSON(w, http.StatusOK, nil)
}
