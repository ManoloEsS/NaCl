package server

import "net/http"

func (s *Server) handlerIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("this is root"))
}
