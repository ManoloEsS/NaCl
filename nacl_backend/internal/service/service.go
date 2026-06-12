package service

import (
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
)

type Service struct {
	Db     *db.Database
	Config *config.Config
}

func New(database *db.Database, cfg *config.Config) *Service {
	return &Service{
		Db:     database,
		Config: cfg,
	}
}
