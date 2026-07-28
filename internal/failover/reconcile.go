package failover

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

func ReconcileFailoverMultiInstance(ctx context.Context, k8sClient client.Client, pods *corev1.PodList, updateStatusFunc UpdateStatusFunc) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	leaderElectionState := GetLeaderElectionState(pods)
	currentLeader := leaderElectionState.CurrentLeader
	hasHealthyLeader := currentLeader != nil

	if hasHealthyLeader {
		if err := updateStatusFunc(&FailoverResult{Leader: currentLeader}); err != nil {
			log.Error(err, "failed to update PiHole cluster status")
			return ctrl.Result{}, err
		}

		// log.V(1).Info("current elected leader", "leader", currentLeader.Name)

		// Dubbel check to demote other pods to failover
		if err := HandleStandbyPodsWithoutRole(ctx, k8sClient, leaderElectionState, nil); err != nil {
			log.Error(err, "could not finish labeling unlabeled standby pods")
			return ctrl.Result{}, err

		}

		return ctrl.Result{}, nil
	}

	// handle startup cases where pods are still getting ready
	if len(leaderElectionState.AvailableLeaderCandidates.Items) == 0 {
		// status ReasonNoEligibleLeader
		return ctrl.Result{RequeueAfter: 5 * time.Second}, nil
	}

	failoverResult, err := Failover(ctx, k8sClient, leaderElectionState)
	if err != nil {
		log.Error(err, "something went wrong during the failover process")
		return ctrl.Result{}, err
	}

	if err := updateStatusFunc(&failoverResult); err != nil {
		log.Error(err, "failed to update PiHole cluster leader status")
		return ctrl.Result{}, err
	}

	log.V(1).Info("failover completed", "leader", failoverResult.Leader.Name)

	return ctrl.Result{}, nil
}

func ReconcileFailoverSingleInstance(ctx context.Context, k8sClient client.Client, pod *corev1.Pod, updateStatusFunc UpdateStatusFunc) (ctrl.Result, error) {
	if !isPodHealthy(pod) {
		return ctrl.Result{
			RequeueAfter: 5 * time.Second,
		}, nil
	}

	if !podIsLeader(pod) {
		if err := promoteToLeader(ctx, k8sClient, pod); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
