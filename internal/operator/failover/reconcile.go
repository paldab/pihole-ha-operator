package failover

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileFailoverMultiInstance(ctx context.Context, k8sClient client.Client, pods *corev1.PodList) (FailoverResult, error) {
	log := logf.FromContext(ctx)

	leaderElectionState := GetLeaderElectionState(pods)
	currentLeader := leaderElectionState.CurrentLeader
	hasHealthyLeader := currentLeader != nil

	if hasHealthyLeader {
		// Dubbel check to demote other pods to failover
		if err := HandleStandbyPodsWithoutRole(ctx, k8sClient, leaderElectionState, nil); err != nil {
			log.Error(err, "could not finish labeling unlabeled standby pods")
			return FailoverResult{}, err
		}

		return FailoverResult{
			InProgress: false,
			Leader:     currentLeader,
			Reason:     ReasonLeaderHealthy,
		}, nil
	}

	return Failover(ctx, k8sClient, leaderElectionState)
}

func ReconcileFailoverSingleInstance(ctx context.Context, k8sClient client.Client, pod *corev1.Pod) (FailoverResult, error) {
	if !isPodHealthy(pod) {
		return FailoverResult{
			InProgress: true,
			Leader:     nil,
			Reason:     ReasonLeaderUnavailable,
		}, nil
	}

	if !podIsLeader(pod) {
		if err := promoteToLeader(ctx, k8sClient, pod); err != nil {
			return FailoverResult{
				InProgress: false,
				Leader:     nil,
				Reason:     ReasonPromotionFailed,
			}, err
		}
	}

	return FailoverResult{
		InProgress: false,
		Leader:     pod,
		Reason:     ReasonLeaderHealthy,
	}, nil
}
