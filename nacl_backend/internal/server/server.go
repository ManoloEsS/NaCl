package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/database"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Config     *config.Config
	HttpServer *http.Server
	Db         *database.Database
	Logger     *slog.Logger
}

func NewServer(
	db *database.Database,
	logger *slog.Logger,
	config *config.Config,
) *Server {
	r := chi.NewRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: r,
	}

	s := &Server{
		Config:     config,
		HttpServer: srv,
		Db:         db,
		Logger:     logger,
	}

	return s
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.HttpServer.Addr)
	if err != nil {
		return err
	}

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("could not find address")
	}
	s.Logger.Debug(fmt.Sprintf("Salt is running on http://localhost:%d", tcpAddr.Port))

	if err := s.HttpServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.HttpServer.Shutdown(ctx)
}

// func main() {
// 	r := chi.NewRouter()
//
// 	// set middleware
// 	// r.Use(middleware.SomeMiddleware)
//
// 	//index routes
// 	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 		w.Write([]byte("root."))
// 	})
//
// 	//create subroutes
// 	//r.Route("/users", func(r chi.Router) {
// 	//r.With(somemiddleware).Get("/", ListUsers)
// 	//})
//
// 	//generate docs
// 	// if *routes {
// 	// 	// fmt.Println(docgen.JSONRoutesDoc(r))
// 	// 	fmt.Println(docgen.MarkdownRoutesDoc(r, docgen.MarkdownOpts{
// 	// 		ProjectPath: "github.com/go-chi/chi/v5",
// 	// 		Intro:       "Welcome to the chi/_examples/rest generated docs.",
// 	// 	}))
// 	// 	return
// 	// }
//
// 	http.ListenAndServe(":3333", r)
// }
