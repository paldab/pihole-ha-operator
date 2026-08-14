// Package envs contain functions for managing env variables
package envs

import (
	"errors"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	exporterapi "github.com/paldab/pihole-ha-operator/api/exporter_api"
	"github.com/paldab/pihole-ha-operator/internal/statistics_exporter/database"
)

func GetIntEnvironment(env string, defaultValue int) int {
	envName := os.Getenv(env)
	if envName == "" {
		return defaultValue
	}

	parsedEnvToInt, err := strconv.Atoi(envName)
	if err != nil {
		log.Printf("could not convert %s with value: %s to int. Using default %d\n", env, envName, defaultValue)
		return defaultValue
	}

	return parsedEnvToInt
}

func GetPiholeFTLDBPath(defaultFTLDBPath string) string {
	ftlPath := os.Getenv("PIHOLE_FTL_DB")
	if ftlPath != "" {
		return ftlPath
	}

	return defaultFTLDBPath
}

func GetDatabaseConfigFromEnvs() database.PostgresConnConfig {
	requiredDBEnvs := map[string]string{
		"DB_HOST":     "",
		"DB_USER":     "",
		"DB_PASSWORD": "",
	}

	var databaseConnection = database.PostgresConnConfig{
		Database: "pihole_statistics",
	}

	for key := range requiredDBEnvs {
		tmpEnv := os.Getenv(key)

		if tmpEnv == "" {
			log.Fatalf("cannot find environment variable '%s'", key)
		}

		requiredDBEnvs[key] = tmpEnv
	}

	dbPort := GetIntEnvironment("DB_PORT", 5432)
	dbName := os.Getenv("DB_DATABASE")

	if dbName != "" {
		databaseConnection.Database = dbName
	}

	databaseConnection.Host = requiredDBEnvs["DB_HOST"]
	databaseConnection.Port = uint16(dbPort)
	databaseConnection.User = requiredDBEnvs["DB_USER"]
	databaseConnection.Password = requiredDBEnvs["DB_PASSWORD"]

	return databaseConnection
}

func GetExporterEnvironments(defaultBatchSize, defaultInterval int) exporterapi.ExporterConfig {
	exportBatchSize := GetIntEnvironment("EXPORTER_BATCH_SIZE", defaultBatchSize)
	exportInterval := GetIntEnvironment("EXPORTER_INTERVAL", defaultInterval)

	return exporterapi.ExporterConfig{
		Batchsize: exportBatchSize,
		Interval:  exportInterval,
	}
}

func GetIdentifierEnvironments() exporterapi.DatabaseIdentifier {
	clusterUUID := os.Getenv("CLUSTER_UUID")
	if clusterUUID == "" {
		log.Fatal("cannot find environment variable 'CLUSTER_UUID'")
	}

	// redo by getting file on path or generate
	sourceIDPath := os.Getenv("SOURCE_ID_DIR")
	if sourceIDPath == "" {
		log.Fatal("cannot find environment variable 'SOURCE_ID_DIR'")
	}

	id, err := getOrCreateSourceID(sourceIDPath)
	if err != nil {
		log.Fatalf(
			"could not fetch or create source ID from %s. err: %v",
			sourceIDPath,
			err,
		)
	}

	return exporterapi.DatabaseIdentifier{ClusterUUID: clusterUUID, SourceUUID: id}
}

func getOrCreateSourceID(dirPath string) (string, error) {
	statisticsExportSourceIDFile := "source-id"
	filePath := dirPath + string(os.PathSeparator) + statisticsExportSourceIDFile
	data, err := os.ReadFile(filePath)

	if err == nil {
		return strings.TrimSpace(string(data)), nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return "", err
	}

	id := uuid.NewString()

	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	if err := os.WriteFile(filePath, []byte(id), 0600); err != nil {
		return "", err
	}

	return id, nil
}
