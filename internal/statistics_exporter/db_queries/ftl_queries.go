// Package dbqueries contain database query functions
package dbqueries

import (
	"database/sql"

	exporterapi "github.com/paldab/pihole-ha-operator/api/exporter_api"
)

// GetLastFTLQuery gets the highest ID in the db.
// returns 0 if no record present
func GetLastFTLQuery(db *sql.DB) (int, error) {
	query := `
	SELECT COALESCE(MAX(id), 0) from queries;
	`

	row := db.QueryRow(query)
	var lastRecordID int

	err := row.Scan(&lastRecordID)

	if err != nil {
		return 0, err
	}

	return lastRecordID, nil
}

func GetPiholeFTLQueries(db *sql.DB, checkpoint int) ([]exporterapi.PiholeFTLQuery, error) {
	var ftlQueries = []exporterapi.PiholeFTLQuery{}
	query := `
		SELECT id, timestamp, type, status, client, domain, reply_time from queries
		WHERE id > $1 
		ORDER BY id ASC;
	`

	rows, err := db.Query(query, checkpoint)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var ftlQuery = exporterapi.PiholeFTLQuery{}

		err := rows.Scan(
			&ftlQuery.ID,
			&ftlQuery.Timestamp,
			&ftlQuery.Type,
			&ftlQuery.Status,
			&ftlQuery.Client,
			&ftlQuery.Domain,
			&ftlQuery.ReplyTime,
		)

		if err != nil {
			return nil, err
		}

		ftlQueries = append(ftlQueries, ftlQuery)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return ftlQueries, nil
}
