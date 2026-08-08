package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	exporterapi "github.com/paldab/pihole-ha-operator/api/exporter_api"
	"github.com/paldab/pihole-ha-operator/internal/statistics_exporter/database"
	dbqueries "github.com/paldab/pihole-ha-operator/internal/statistics_exporter/db_queries"
	"github.com/paldab/pihole-ha-operator/internal/statistics_exporter/envs"
	_ "modernc.org/sqlite"
)

const (
	defaultBatchSize = 100000
	defaultInterval  = 60
	defaultFTLDBPath = "/etc/pihole/pihole-FTL.db"

	migrationDirectory = "internal/statistics_exporter/migrations"
)

var (
	exporterIdentifier = exporterapi.DatabaseIdentifier{}
	exporterConfig     = exporterapi.ExporterConfig{}
)

func main() {
	exporterIdentifier = envs.GetK8sEnvironments()
	exporterConfig = envs.GetExporterEnvironments(defaultBatchSize, defaultInterval)
	externalDatabaseConfig := envs.GetDatabaseConfigFromEnvs()

	ftlDBPath := envs.GetPiholeFTLDBPath(defaultFTLDBPath)
	piholeLocalDBFile := fmt.Sprintf("file:%s?mode=ro", ftlDBPath)
	localDB, err := database.NewDB("sqlite", piholeLocalDBFile)

	if err != nil {
		log.Fatalf("could not connect to the pihole sqlite database on path %s, err: %v", ftlDBPath, err)
	}
	defer localDB.Close() //nolint:errcheck

	externalPostgresConnString := database.CreatePostgresConnString(externalDatabaseConfig)
	migrationDBConn, err := database.NewDB("pgx", externalPostgresConnString)

	if err != nil {
		log.Fatal(err)
	}

	if err := database.RunMigrations(migrationDBConn, migrationDirectory); err != nil {
		log.Fatal(err)
	}
	migrationDBConn.Close() //nolint:errcheck

	ctx := context.Background()
	pool, err := database.NewPostgresPool(ctx, externalDatabaseConfig)

	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	startExporterLoop(ctx, localDB, pool)
}

func startExporterLoop(ctx context.Context, db *sql.DB, pool *pgxpool.Pool) {
	interval := time.Second * time.Duration(exporterConfig.Interval)
	ticker := time.NewTicker(interval)
	quit := make(chan struct{})

	log.Printf(
		"starting pihole statistics exporter (source_pvc=%s, pvc_id=%s, cluster_id=%s)\n",
		exporterIdentifier.SourcePVCName,
		exporterIdentifier.SourcePVCUUID,
		exporterIdentifier.ClusterUUID,
	)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := exportQueries(ctx, db, pool); err != nil {
				log.Printf("export failed: %v", err)
			}

		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func exportQueries(ctx context.Context, db *sql.DB, pool *pgxpool.Pool) error {
	mostRecentPiholeFTLQueryID, err := dbqueries.GetLastFTLQuery(db)

	if err != nil {
		return err
	}

	// pihole doesnt have any queries yet
	if mostRecentPiholeFTLQueryID == 0 {
		return nil
	}

	var lastCheckpointResult = dbqueries.GetLatestCheckpoint(ctx, pool, exporterIdentifier)
	var lastCheckpointID = 0

	if lastCheckpointResult != nil {
		log.Printf(
			"found %d new entries compared with the last checkpoint",
			(mostRecentPiholeFTLQueryID - int(lastCheckpointResult.LastExportedID)),
		)

		lastCheckpointID = int(lastCheckpointResult.LastExportedID)
	}

	newFTLRecordsPresent := mostRecentPiholeFTLQueryID != lastCheckpointID
	if newFTLRecordsPresent {
		queryResults, err := dbqueries.GetPiholeFTLQueries(db, lastCheckpointID)

		if err != nil {
			return fmt.Errorf("failed to get pihole ftl queries: %v", err)
		}

		queryResultsLen := len(queryResults)

		if queryResultsLen == 0 {
			return fmt.Errorf("found 0 query results when the checkpoint id was the same as the last pihole query statistic")
		}

		if queryResultsLen >= exporterConfig.Batchsize {
			log.Printf(
				"found more queries than the batchSize missing in the external database. Inserting %d queries with a batchsize: %d",
				queryResultsLen, exporterConfig.Batchsize,
			)
		}

		for processedQueries := 0; processedQueries < queryResultsLen; {
			// as long as the batch doesnt exceed the queryResultsLen, choose batch otherwise queryResultsLen
			end := min(processedQueries+exporterConfig.Batchsize, queryResultsLen)
			batch := queryResults[processedQueries:end]

			if err := dbqueries.InsertQueries(ctx, pool, batch, exporterIdentifier); err != nil {
				return fmt.Errorf("failed to insert queries in the external db: %v", err)
			}

			processedQueries = end
		}

		log.Default().Printf("Exported %d queries to the external db\n", queryResultsLen)
	}

	return nil
}
