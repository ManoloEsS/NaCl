package server

import (
	"io"
	"io/fs"
	"net/http"
)

func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request) {
	root, _ := fs.Sub(s.StaticFS, "dist")
	index, err := root.Open("index.html")
	if err != nil {
		http.Error(w, "Not Found", http.StatusNotFound)
		s.Logger.Error("could not serve index.html")
		return
	}
	defer index.Close()
	w.Header().Set("Content-Type", "text/html")
	_, _ = io.Copy(w, index)
}
