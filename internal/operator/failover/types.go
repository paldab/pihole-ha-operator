package failover

const (
	ReasonLeaderUnavailable = "LeaderUnavailable"
	ReasonFailoverComplete  = "FailoverComplete"
	ReasonLeaderUnhealthy   = "LeaderUnhealthy"
	ReasonPromotionFailed   = "LeaderPromotionFailed"
	ReasonDemotionFailed    = "PodDemotionFailed"
	ReasonLeaderHealthy     = "LeaderHealthy"
)
