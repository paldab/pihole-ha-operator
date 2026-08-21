package database

import (
	"errors"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func IsSQLiteBusy(err error) bool {
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
