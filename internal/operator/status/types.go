// Package status contain all objects that are relevant to the status of the crds
package status

const (
	TypeReady       = "Ready"
	TypeFailingOver = "FailoverInProgress"
)

const (
	ConditionClusterReady = "Ready"
	ConditionFailingOver  = "FailingOver"
)

const (
	ReasonClusterHealthy      = "ClusterHealthy"
	ReasonStatefulSetNotReady = "StatefulSetNotReady"
	ReasonStatefulSetUpdating = "StatefulSetUpdating"
	ReasonLeaderUnavailable   = "LeaderUnavailable"
	ReasonFailoverInProgress  = "FailoverInProgress"
	ReasonFailoverCompleted   = "FailoverCompleted"
)

// PiholeConfig
const (
	ConditionConfigReady = "Ready"

	ReasonConfigurationApplied = "ConfigurationApplied"
	ReasonDuplicateClusterRef  = "DuplicateClusterRef"
	ReasonClusterNotFound      = "ClusterNotFound"
	ReasonReconcileFailed      = "ReconcileFailed"
)
