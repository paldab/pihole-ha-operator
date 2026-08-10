package defaults

import "github.com/paldab/pihole-ha-operator/api/v1alpha1"

const (
	StatisticsExporterImage   = "paldab.nl/pihole-ha-statistics-exporter"
	StatisticsMode            = v1alpha1.StatsModeLocal
	StatisticsExportInterval  = 60
	StatisticsExportBatchSize = 100000
)
