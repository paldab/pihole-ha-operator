package dbqueries

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	exporterapi "github.com/paldab/pihole-ha-operator/api/exporter_api"
)

func GetLatestCheckpoint(
	ctx context.Context,
	pool *pgxpool.Pool,
	dbIdentifier exporterapi.DatabaseIdentifier,
) *exporterapi.PiholeDatabaseCheckpoint {
	query := `
		SELECT last_exported_id FROM pihole_database_checkpoint
		WHERE cluster_uuid = $1
		AND source_uuid = $2
	`

	row := pool.QueryRow(ctx, query, dbIdentifier.ClusterUUID, dbIdentifier.SourceUUID)

	var checkpoint exporterapi.PiholeDatabaseCheckpoint
	if err := row.Scan(&checkpoint.LastExportedID); err != nil {
		// no rows found
		return nil
	}

	return &checkpoint
}

func UpdateCheckpoint(
	ctx context.Context,
	tx pgx.Tx,
	dbIdentifier exporterapi.DatabaseIdentifier,
	lastQueryID int,
) error {
	query := `
	INSERT INTO pihole_database_checkpoint (
		cluster_uuid,
		source_uuid,
		last_exported_id
	)
	VALUES ($1,$2,$3)
	ON CONFLICT (
		cluster_uuid,
		source_uuid
	)
	DO UPDATE SET
		last_exported_id = EXCLUDED.last_exported_id;
	`

	_, err := tx.Exec(ctx, query, dbIdentifier.ClusterUUID, dbIdentifier.SourceUUID, lastQueryID)

	return err
}

// InsertQueries inserts exported queries to the external database
func InsertQueries(
	ctx context.Context,
	pool *pgxpool.Pool,
	queries []exporterapi.PiholeFTLQuery,
	dbIdentifier exporterapi.DatabaseIdentifier,
) error {
	tableIdentifier := pgx.Identifier{"pihole_queries"}
	rows := make([][]any, 0, len(queries))
	piholeQueryCols := []string{
		"cluster_uuid",
		"source_uuid",
		"local_query_id",
		"query_time",
		"type",
		"status",
		"client",
		"domain",
		"reply_time",
	}

	for _, ftlQuery := range queries {
		rows = append(rows, []any{
			dbIdentifier.ClusterUUID,
			dbIdentifier.SourceUUID,
			ftlQuery.ID,
			ftlQuery.Timestamp,
			exporterapi.PiholeQueryType(ftlQuery.Type).String(),
			exporterapi.PiholeQueryStatus(ftlQuery.Status).String(),
			ftlQuery.Client,
			ftlQuery.Domain,
			ftlQuery.ReplyTime,
		})
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("something went wrong when starting a transaction: %v", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	_, err = tx.CopyFrom(ctx, tableIdentifier, piholeQueryCols, pgx.CopyFromRows(rows))
	if err != nil {
		return fmt.Errorf("something went wrong when copying from rows: %v", err)
	}

	lastCheckpointID := queries[len(queries)-1].ID

	err = UpdateCheckpoint(
		ctx,
		tx,
		dbIdentifier,
		int(lastCheckpointID),
	)

	if err != nil {
		return fmt.Errorf("could not update checkpoint: %v", err)
	}

	return tx.Commit(ctx)
}
