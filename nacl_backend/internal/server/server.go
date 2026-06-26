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
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/service"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	Config     *config.Config
	HTTPServer *http.Server
	Svc        *service.Service
	Logger     *slog.Logger
}

func NewServer(
	db db.Querier,
	logger *slog.Logger,
	config *config.Config,
) *Server {
	r := chi.NewRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: r,
	}

	svc := service.New(db, config)

	s := &Server{
		Config:     config,
		HTTPServer: srv,
		Svc:        svc,
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

	r.Get("/", s.HandleIndex)

	r.Post("/api/users", s.HandleCreateUser)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Patch("/api/users", s.HandleUpdateUserPassword)

	r.Post("/api/login", s.HandleLogin)

	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Post("/api/credentials", s.HandleCreateCredential)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Get("/api/credentials", s.HandleListCredentials)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Post("/api/credentials/{id}/decrypt", s.HandleDecryptCredentialByID)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Patch("/api/credentials/{id}", s.HandleUpdateCredentialPassword)
	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Delete("/api/credentials/{id}", s.HandleDeleteCredential)

	r.With(middleware.TokenValidator(s.Logger, s.Config.JwtSecret)).
		Get("/api/operations", s.HandleListOpsforUserID)
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
