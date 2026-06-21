package server

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/auth"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
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

	s := &Server{
		Config: cfg,
		Svc:    service.New(database, cfg),
		Logger: logger,
	}

	r := chi.NewRouter()
	s.RegisterRoutes(r)

	s.HTTPServer = &http.Server{Handler: r}

	return s
}

func createTestUser(t *testing.T, database *db.Database, username, password string) {
	t.Helper()

	passHash, err := auth.HashPassword(password)
	require.NoError(t, err)

	salt, err := encryption.GenerateRandomBytes(32)
	require.NoError(t, err)

	key, err := encryption.GenerateRandomBytes(32)
	require.NoError(t, err)

	derivedKey, err := encryption.DeriveKey(password, salt)
	require.NoError(t, err)

	encryptedMasterKey, err := encryption.Encrypt(key, derivedKey)
	require.NoError(t, err)

	err = database.Queries().CreateUser(context.Background(), db.CreateUserParams{
		Username:           username,
		PasswordHash:       passHash,
		MasterKeySalt:      base64.StdEncoding.EncodeToString(salt),
		EncryptedMasterKey: base64.StdEncoding.EncodeToString(encryptedMasterKey),
	})
	require.NoError(t, err)
}

func loginTestUser(t *testing.T, database *db.Database, jwtSecret, username, password string) string {
	t.Helper()
	createTestUser(t, database, username, password)

	user, err := database.Queries().GetUserByUsername(context.Background(), username)
	require.NoError(t, err, "loginTestUser: user not found after creation")

	token, err := auth.MakeJWT(user.ID, jwtSecret, 30*time.Minute)
	require.NoError(t, err, "loginTestUser: failed to create JWT")
	return token
}
