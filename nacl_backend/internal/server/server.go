package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Config     *config.Config
	HTTPServer *http.Server
	Db         *db.Database
	Logger     *slog.Logger
}

func NewServer(
	db *db.Database,
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
		HTTPServer: srv,
		Db:         db,
		Logger:     logger,
	}

	s.RegisterRoutes(r)

	return s
}

func (s *Server) RegisterRoutes(r chi.Router) {
	r.Use(
		middleware.RequestLogger(s.Logger),
		middleware.Recovery(s.Logger),
	)

	r.Get("/", s.handlerIndex)

	r.Post("/api/users", s.handlerCreateUser)

	r.Post("/api/login", s.handlerLogin)

	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Post("/api/services", s.handlerCreateService)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Get("/api/services", s.handlerGetAllServicesUser)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Post("/api/services/{id}/credentials", s.handlerDecryptById)
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.HTTPServer.Addr)
	if err != nil {
		return err
	}

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("could not find address")
	}
	s.Logger.Debug(fmt.Sprintf("Salt is running on http://localhost:%d", tcpAddr.Port))

	if err := s.HTTPServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.HTTPServer.Shutdown(ctx)
}
