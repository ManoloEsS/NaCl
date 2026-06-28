package config

import (
	"encoding/base64"
	"fmt"
	"os"
	"strconv"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/encryption"
	"github.com/joho/godotenv"
)

type Config struct {
	JwtSecret string
	Port      int
	LogFile   string
	DbString  string
}

func NewConfig() (*Config, error) {
	var httpPort int
	var secret string
	var logFile string
	var dbString string

	_ = godotenv.Load()

	saltPort := os.Getenv("SALT_PORT")
	if saltPort == "" {
		saltPort = os.Getenv("PORT")
		fmt.Fprintf(os.Stderr, "invalid SALT_PORT, using fallback port %v\n", httpPort)
	}
	httpPort, _ = strconv.Atoi(saltPort)

	secret = os.Getenv("JWT_SECRET")
	if secret == "" {
		randomBytes, _ := encryption.GenerateRandomBytes(16)
		secret = base64.StdEncoding.EncodeToString(randomBytes)
		fmt.Fprintf(os.Stderr, "invalid secret, generating random secret")
	}

	logFile = os.Getenv("SALT_LOG_FILE")
	if logFile == "" {
		logFile = "../../salt.access.log"
	}

	dbString = os.Getenv("DATABASE_URL")
	if dbString == "" {
		return nil, fmt.Errorf("no database string found: exiting with error")
	}

	cfg := &Config{
		JwtSecret: secret,
		LogFile:   logFile,
		Port:      httpPort,
		DbString:  dbString,
	}

	return cfg, nil
}
