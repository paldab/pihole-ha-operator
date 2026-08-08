package database

import "database/sql"

func NewDB(driver, connectionString string) (*sql.DB, error) {
	db, err := sql.Open(driver, connectionString)

	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
