// Package status contain all objects that are relevant to the status of the crds
package status

const (
	TypeReady              = "Ready"
	TypeFailoverInProgress = "FailoverInProgress"
)

const (
	ReasonNoEligibleLeader = "NoEligibleLeader"
	ReasonFailoverComplete = "FailoverComplete"
	ReasonLeaderUnhealthy  = "LeaderUnhealthy"
	ReasonPromotionFailed  = "LeaderPromotionFailed"
)

var FailoverReasonMessages = map[string]string{
	ReasonNoEligibleLeader: "There is not a single pod stable enough to become a leader",
	ReasonFailoverComplete: "Failover comeplete",
	ReasonLeaderUnhealthy:  "Leader has become unhealthy and can't receive traffic",
}
