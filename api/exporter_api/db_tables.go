// Package exporterapi contain types for the statistics exporter
package exporterapi

import "database/sql"

type DatabaseIdentifier struct {
	ClusterUUID   string `db:"cluster_uuid"`    // Pihole cluster ID
	SourcePVCUUID string `db:"source_pvc_uuid"` // ID of the PV
	SourcePVCName string `db:"source_pvc_name"` // PV like pihole-storage-0
}

type PiholeDatabaseCheckpoint struct {
	DatabaseIdentifier

	LastExportedID int64 `db:"last_exported_id"` // Highest local Pi-hole query ID successfully exported.
}

type PiholeFTLQuery struct {
	ID        int64           `db:"id"`
	Timestamp float64         `db:"timestamp"`
	Type      int             `db:"type"`
	Status    int             `db:"status"` // Status like denied or allowed
	Client    string          `db:"client"` // IP from device trying to connect to domain
	Domain    string          `db:"domain"` // Domain trying to reach
	ReplyTime sql.NullFloat64 `db:"reply_time"`
}

// delayed cols
// ReplyType      int
// Forward      sql.NullString
// ListID       sql.NullInt64
// EDE          int
