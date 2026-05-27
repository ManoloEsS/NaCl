package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	JwtSecret string
	Port      int
	LogFile   string
	DbString  string
}

func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load env vars")
		return nil, err
	}

	httpPort, err := strconv.Atoi(os.Getenv("SALT_PORT"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid SALT_PORT : %v\n", err)
		return nil, err
	}

	cfg := &Config{
		JwtSecret: os.Getenv("JWT_SECRET"),
		LogFile:   os.Getenv("SALT_LOG_FILE"),
		Port:      httpPort,
		DbString:  os.Getenv("DATABASE_URL"),
	}

	return cfg, nil
}
