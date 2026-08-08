// Package envs contain functions for managing env variables
package envs

import (
	"log"
	"os"
	"strconv"

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

func GetK8sEnvironments() exporterapi.DatabaseIdentifier {
	clusterUUID := os.Getenv("CLUSTER_UUID")
	if clusterUUID == "" {
		log.Fatal("cannot find environment variable 'CLUSTER_UUID'")
	}

	pvcUUID := os.Getenv("PVC_UUID")
	if pvcUUID == "" {
		log.Fatal("cannot find environment variable 'PVC_UUID'")
	}

	pvcName := os.Getenv("PVC_NAME")
	if pvcName == "" {
		log.Fatal("cannot find environment variable 'PVC_NAME'")
	}

	return exporterapi.DatabaseIdentifier{ClusterUUID: clusterUUID, SourcePVCUUID: pvcUUID, SourcePVCName: pvcName}
}
