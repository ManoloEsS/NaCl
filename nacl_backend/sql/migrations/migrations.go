package migrations

import (
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var MigrationFS embed.FS

func RunMigrations(pool *pgxpool.Pool) error {
	goose.SetBaseFS(MigrationFS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	db := stdlib.OpenDBFromPool(pool)
	if err := goose.Up(db, "."); err != nil {
		return err
	}
	return db.Close()
}
