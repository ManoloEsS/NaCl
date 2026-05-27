package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type server struct {
	httpServer *http.Server
	cancel     context.CancelFunc
	logger     *slog.Logger
}

func NewServer(port int, cancel context.CancelFunc, logger *slog.Logger) *server {
	r := chi.NewRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	s := &server{
		httpServer: srv,
		cancel:     cancel,
		logger:     logger,
	}

	return s
}

func (s *server) Start() error {
	ln, err := net.Listen("tcp", s.httpServer.Addr)
	if err != nil {
		return err
	}

	tcpAddr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("could not find address")
	}
	s.logger.Debug(fmt.Sprintf("Salt is running on http://localhost:%d", tcpAddr.Port))

	if err := s.httpServer.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// func (s *server) handlerShutdown(w http.ResponseWriter, r *http.Request) {
// 	if os.Getenv("ENV") == "production" {
// 		http.NotFound(w, r)
// 		return
// 	}
// 	w.WriteHeader(http.StatusOK)
// 	go s.cancel()
// }
//
// func requestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			next.ServeHTTP(w, r)
// 			logger.Info("Served request",
// 				slog.String("method", r.Method),
// 				slog.String("path", r.URL.Path),
// 				slog.String("client_ip", r.RemoteAddr),
// 			)
// 		})
// 	}
// }
