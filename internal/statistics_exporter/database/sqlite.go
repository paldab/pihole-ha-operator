package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func WaitForFTLDatabasePresent(path string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if _, err := os.Stat(path); err == nil {
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for FTL database %q", path)
		}

		time.Sleep(2 * time.Second)
	}
}

func WaitForFTLDatabaseNotBusy(path string, timeout time.Duration) (*sql.DB, error) {
	deadline := time.Now().Add(2 * time.Minute)

	for {
		localDB, err := NewDB("sqlite", path)

		if err == nil {
			return localDB, nil
		}

		if isSQLiteBusy(err) {
			if time.Now().After(deadline) {
				return nil, fmt.Errorf(
					"pihole database remained locked for more than 2 minutes: %v",
					err,
				)
			}

			log.Printf("pihole database is locked, retrying in 15 seconds")
			time.Sleep(15 * time.Second)
			continue
		}
	}
}

func isSQLiteBusy(err error) bool {
	if err == nil {
		return false
	}

	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	// SQLite extended result codes keep the primary error code
	// in the low 8 bits.
	code := sqliteErr.Code() & 0xff

	return code == sqlite3.SQLITE_BUSY ||
		code == sqlite3.SQLITE_LOCKED
}
