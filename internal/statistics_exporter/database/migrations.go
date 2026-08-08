package database

import (
	"database/sql"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB, migrationDir string) error {
	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return err
	}

	return goose.Up(db, migrationDir)
}
