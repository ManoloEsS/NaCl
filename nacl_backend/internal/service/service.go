package service

import (
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/config"
	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
)

type Service struct {
	Queries db.Querier
	Config  *config.Config
}

func New(database db.Querier, cfg *config.Config) *Service {
	return &Service{
		Queries: database,
		Config:  cfg,
	}
}
