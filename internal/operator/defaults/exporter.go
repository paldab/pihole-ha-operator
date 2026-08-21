package defaults

import "github.com/paldab/pihole-ha-operator/api/v1alpha1"

const (
	StatisticsExporterImage   = "ghcr.io/paldab/pihole-ha-statistics-exporter"
	StatisticsMode            = v1alpha1.StatsModeLocal
	StatisticsExportInterval  = 60
	StatisticsExportBatchSize = 100000

	StatisticsExportSourceIDDir  = "/etc/pihole/.pihole-ha"
	StatisticsExportSourceIDFile = "source-id"
)
