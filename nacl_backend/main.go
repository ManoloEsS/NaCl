package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/logger"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	status := run(ctx)
	cancel()
	fmt.Fprintf(os.Stderr, "shutting down completely with status %v\n", status)
	os.Exit(status)
}

func run(ctx context.Context) int {
	fmt.Fprint(os.Stderr, "loading config\n")
	cfg, err := config.NewConfig()
	if err != nil {
		return 1
	}
	fmt.Fprint(os.Stderr, "config loaded successfully\n")

	fmt.Fprint(os.Stderr, "initializing logger\n")
	log, closeLogger, err := logger.InitializeLogger(cfg.LogFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		return 1
	}
	log.Debug("logger succesfully initialized")

	defer func() {
		log.Debug("closing logger")
		if err := closeLogger(); err != nil {
			fmt.Fprintf(os.Stderr, "logger shutdown with error: %v\n", err)
		}
		fmt.Fprint(os.Stderr, "logger closed successfully\n")
	}()

	pool, err := pgxpool.New(ctx, cfg.DbString)
	if err != nil {
		log.Error("failes to connect to db", "error", err)
		return 1
	}
	defer pool.Close()

	queries := db.New(pool)

	s := server.NewServer(queries, log, cfg, DistFS)

	var serverErr error
	go func() {
		log.Debug("starting server in go routine")
		serverErr = s.Start()
		log.Debug("server shutting down")
	}()

	// blocks
	<-ctx.Done()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	log.Debug("Salt is shutting down")
	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to shutdown server", "error", err)
		return 1
	}
	log.Debug("server shut down")
	if serverErr != nil {
		log.Error("server error", "error", serverErr)
		return 1
	}
	log.Debug("server shut down with no errors")
	return 0
}
