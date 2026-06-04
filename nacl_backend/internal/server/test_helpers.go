package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
)

func cleanupTestDB(t *testing.T, database *db.Database, tables ...string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	truncateString := fmt.Sprintf("TRUNCATE %s CASCADE", strings.Join(tables, ", "))
	_, err := database.Pool.Exec(ctx, truncateString)
	if err != nil {
		t.Fatalf("failed to cleanup test database: %v", err)
	}

}

func newTestDB(t *testing.T) *db.Database {
	t.Helper()

	dbURL := os.Getenv("DATABASE_URL_TEST")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database, err := db.NewDatabase(ctx, dbURL)
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	return database
}

func newTestServer(t *testing.T, database *db.Database) *Server {
	t.Helper()

	cfg := &config.Config{
		Port:      3334,
		JwtSecret: "test-secret",
		LogFile:   "/tmp/test.log",
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	return &Server{
		Config: cfg,
		Db:     database,
		Logger: logger,
	}
}
