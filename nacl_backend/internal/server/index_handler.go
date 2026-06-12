package server

import "net/http"

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte("this is root"))
}
