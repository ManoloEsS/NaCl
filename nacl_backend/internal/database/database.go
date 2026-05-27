package database

import (
	"context"

	"github.com/ManoloEsS/NaCl/nacl_backend/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	Pool *pgxpool.Pool
}

func NewDatabase(ctx context.Context, connString string) (*Database, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return &Database{Pool: pool}, nil
}

func (d *Database) Close() {
	d.Pool.Close()
}

func (d *Database) Queries() *db.Queries {
	return db.New(d.Pool)
}

//TODO: add transaction helper for atomic transactions
