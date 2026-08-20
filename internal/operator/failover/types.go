package failover

import corev1 "k8s.io/api/core/v1"

const (
	ReasonLeaderUnavailable = "LeaderUnavailable"
	ReasonFailoverCompleted = "FailoverCompleted"
	ReasonLeaderUnhealthy   = "LeaderUnhealthy"
	ReasonPromotionFailed   = "LeaderPromotionFailed"
	ReasonDemotionFailed    = "PodDemotionFailed"
	ReasonLeaderHealthy     = "LeaderHealthy"
)

type LeaderElectionState struct {
	CurrentLeader             *corev1.Pod
	PreviousLeader            *corev1.Pod
	AvailableLeaderCandidates corev1.PodList
}

type FailoverStatus struct {
	InProgress bool
	LeaderName *string
	Reason     string
	Message    string
}

type FailoverResult struct {
	InProgress bool
	Leader     *corev1.Pod
	Reason     string
}
