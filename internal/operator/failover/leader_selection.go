// Package failover handles the failover to leader or to standby
package failover

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Failover checks if there is a leader currently and promotes one of the pods if there is not a stable leader. Returns error
func Failover(ctx context.Context, k8sClient client.Client, electionState LeaderElectionState) (FailoverResult, error) {
	leaderCandidates := electionState.AvailableLeaderCandidates

	if electionState.PreviousLeader != nil {
		if err := demoteToStandby(ctx, k8sClient, electionState.PreviousLeader); err != nil {
			return FailoverResult{
				InProgress: true,
				Leader:     nil,
				Reason:     ReasonLeaderUnhealthy,
			}, fmt.Errorf("demote previous leader %s: %w", electionState.PreviousLeader.Name, err)
		}
	}

	// handle case where somehow leaderCandidates are still empty
	if len(leaderCandidates.Items) == 0 {
		return FailoverResult{
			InProgress: true,
			Leader:     nil,
			Reason:     ReasonLeaderUnavailable,
		}, fmt.Errorf("there is currently no leader and there are no candidates to become the pihole leader")
	}

	// first available pod and elect it as leader
	firstAvailableLeaderCanidate := &leaderCandidates.Items[0]

	// Create copy of name so that HandleStandbyPodsWithoutRole does not modify the pointer
	newLeaderName := firstAvailableLeaderCanidate.Name
	err := promoteToLeader(ctx, k8sClient, firstAvailableLeaderCanidate)

	if err != nil {
		return FailoverResult{
			InProgress: false,
			Leader:     nil,
			Reason:     ReasonPromotionFailed,
		}, err
	}

	// Make other available pods to standby
	if err = HandleStandbyPodsWithoutRole(ctx, k8sClient, electionState, &newLeaderName); err != nil {
		return FailoverResult{
			InProgress: false,
			Leader:     nil,
			Reason:     ReasonDemotionFailed,
		}, err
	}

	return FailoverResult{
		InProgress: false,
		Leader:     firstAvailableLeaderCanidate,
		Reason:     ReasonFailoverCompleted,
	}, nil
}

// GetLeaderElectionState fetches the current leader pod and the pods ready to become leaders if necessary
// It is possible that the first candidate to become leader is not always the first index of the available pods
// Returns LeaderElectionState
func GetLeaderElectionState(availabePiholePods *corev1.PodList) LeaderElectionState {
	var currentLeader *corev1.Pod = nil
	var previousLeader *corev1.Pod = nil
	var primaryPodCandidates = corev1.PodList{}

	// optional: sort the availabePiholePods

	for idx := range availabePiholePods.Items {
		pod := &availabePiholePods.Items[idx]
		isLeaderPod := podIsLeader(pod)

		if !isPodHealthy(pod) {
			// if pod is already primary but no longer ready to stay primary, reset state
			if isLeaderPod {
				previousLeader = pod
			}

			continue
		}

		// Healthy primary pod found
		if isLeaderPod {
			currentLeader = pod
			break
		}

		// Add as primary candidate
		podCandidate := []corev1.Pod{*pod}
		primaryPodCandidates = corev1.PodList{
			Items: append(primaryPodCandidates.Items, podCandidate...),
		}
	}

	return LeaderElectionState{
		CurrentLeader:             currentLeader,
		PreviousLeader:            previousLeader,
		AvailableLeaderCandidates: primaryPodCandidates,
	}
}

// HandleStandbyPodsWithoutRole makes sure that it labels available non leader pods with the standby labels
func HandleStandbyPodsWithoutRole(ctx context.Context, k8sClient client.Client, electionState LeaderElectionState, newLeaderName *string) error {
	pods := electionState.AvailableLeaderCandidates.Items

	for _, pod := range pods {
		// Don't process the new leader
		if newLeaderName != nil {
			if pod.Name == *newLeaderName {
				continue
			}
		}

		if pod.Labels == nil {
			pod.Labels = map[string]string{}
		}

		if podIsStandby(&pod) {
			continue
		}

		if err := demoteToStandby(ctx, k8sClient, &pod); err != nil {
			return err
		}
	}

	return nil
}
