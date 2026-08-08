package failover_test

import (
	"context"
	"testing"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/failover"
	"github.com/paldab/pihole-ha-operator/internal/operator/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetLeaderElectionState_HealthyPrimaryExists(t *testing.T) {
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			newTestPod("pihole-0", "primary", true),
			newTestPod("pihole-1", "standby", true),
			newTestPod("pihole-2", "standby", true),
		},
	}

	state := failover.GetLeaderElectionState(pods)

	if state.CurrentLeader == nil {
		t.Fatal("expected current leader, got nil")
	}

	if state.CurrentLeader.Name != "pihole-0" {
		t.Fatalf(
			"expected current leader pihole-0, got %s",
			state.CurrentLeader.Name,
		)
	}

	if state.PreviousLeader != nil {
		t.Fatalf(
			"expected no previous leader, got %s",
			state.PreviousLeader.Name,
		)
	}

	// Your implementation stops after finding the healthy primary.
	if len(state.AvailableLeaderCandidates.Items) != 0 {
		t.Fatalf(
			"expected no candidates after finding primary, got %d",
			len(state.AvailableLeaderCandidates.Items),
		)
	}
}

func TestGetLeaderElectionState_UnhealthyPrimaryAndHealthyCandidates(t *testing.T) {
	const primaryPod = "pihole-0"
	const standbyPod = "pihole-1"
	pods := &corev1.PodList{
		Items: []corev1.Pod{
			newTestPod(primaryPod, "primary", false),
			newTestPod(standbyPod, "standby", true),
		},
	}

	state := failover.GetLeaderElectionState(pods)

	if state.CurrentLeader != nil {
		t.Fatalf(
			"expected no current leader, got %s",
			state.CurrentLeader.Name,
		)
	}

	if state.PreviousLeader == nil {
		t.Fatal("expected previous unhealthy leader, got nil")
	}

	if state.PreviousLeader.Name != primaryPod {
		t.Fatalf(
			"expected previous leader %s, got %s",
			primaryPod,
			state.PreviousLeader.Name,
		)
	}

	if len(state.AvailableLeaderCandidates.Items) != 1 {
		t.Fatalf(
			"expected 2 leader candidates, got %d",
			len(state.AvailableLeaderCandidates.Items),
		)
	}

	if state.AvailableLeaderCandidates.Items[0].Name != standbyPod {
		t.Fatalf(
			"expected first candidate %s, got %s",
			standbyPod,
			state.AvailableLeaderCandidates.Items[0].Name,
		)
	}
}

func TestGetLeaderElectionState_ExcludesTerminatingReadyPod(t *testing.T) {
	terminatingPod := newTestPod("pihole-0", "standby", true)
	now := metav1.Now()
	terminatingPod.DeletionTimestamp = &now

	healthyPod := newTestPod("pihole-1", "standby", true)

	pods := &corev1.PodList{
		Items: []corev1.Pod{
			terminatingPod,
			healthyPod,
		},
	}

	state := failover.GetLeaderElectionState(pods)

	if len(state.AvailableLeaderCandidates.Items) != 1 {
		t.Fatalf(
			"expected 1 candidate, got %d",
			len(state.AvailableLeaderCandidates.Items),
		)
	}

	if state.AvailableLeaderCandidates.Items[0].Name != "pihole-1" {
		t.Fatalf(
			"expected candidate pihole-1, got %s",
			state.AvailableLeaderCandidates.Items[0].Name,
		)
	}
}

func TestFailover_DemotesPreviousLeaderAndPromotesFirstCandidate(t *testing.T) {
	ctx := context.Background()

	previousLeader := newTestPod("pihole-0", "primary", false)
	firstCandidate := newTestPod("pihole-1", "standby", true)
	secondCandidate := newTestPod("pihole-2", "", true)

	k8sClient := newFakeClient(
		t,
		previousLeader.DeepCopy(),
		firstCandidate.DeepCopy(),
		secondCandidate.DeepCopy(),
	)

	state := failover.LeaderElectionState{
		PreviousLeader: previousLeader.DeepCopy(),
		AvailableLeaderCandidates: corev1.PodList{
			Items: []corev1.Pod{
				firstCandidate,
				secondCandidate,
			},
		},
	}

	result, err := failover.Failover(
		ctx,
		k8sClient,
		state,
	)

	if err != nil {
		t.Fatalf("Failover returned unexpected error: %v", err)
	}

	if result.InProgress {
		t.Fatal("expected failover to be complete")
	}

	if result.Leader == nil {
		t.Fatal("expected promoted leader, got nil")
	}

	if result.Leader.Name != "pihole-1" {
		t.Fatalf(
			"expected pihole-1 to be promoted, got %s",
			result.Leader.Name,
		)
	}

	if result.Reason != status.ReasonFailoverComplete {
		t.Fatalf(
			"expected reason %q, got %q",
			status.ReasonFailoverComplete,
			result.Reason,
		)
	}

	assertPodRole(t, ctx, k8sClient, "pihole-0", "standby")
	assertPodRole(t, ctx, k8sClient, "pihole-1", "primary")
	assertPodRole(t, ctx, k8sClient, "pihole-2", "standby")
}

func TestFailover_NoEligibleCandidates(t *testing.T) {
	ctx := context.Background()
	k8sClient := newFakeClient(t)

	result, err := failover.Failover(
		ctx,
		k8sClient,
		failover.LeaderElectionState{},
	)

	if err == nil {
		t.Fatal("expected error when no leader candidates are available")
	}

	if result.InProgress {
		t.Fatal("expected InProgress to be false")
	}

	if result.Leader != nil {
		t.Fatalf(
			"expected no leader, got %s",
			result.Leader.Name,
		)
	}

	if result.Reason != status.ReasonNoEligibleLeader {
		t.Fatalf(
			"expected reason %q, got %q",
			status.ReasonNoEligibleLeader,
			result.Reason,
		)
	}
}

func newTestPod(
	name string,
	role string,
	ready bool,
) corev1.Pod {
	labels := map[string]string{}

	if role != "" {
		labels[defaults.RoleLabel] = role
	}

	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}

	return corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: defaultNamespace,
			Labels:    labels,
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: readyStatus,
				},
			},
		},
	}
}

func newFakeClient(
	t *testing.T,
	objects ...client.Object,
) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()

	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core Kubernetes API to scheme: %v", err)
	}

	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add PiHoleCluster API to scheme: %v", err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()
}

func assertPodRole(
	t *testing.T,
	ctx context.Context,
	k8sClient client.Client,
	name string,
	expectedRole string,
) {
	t.Helper()

	pod := &corev1.Pod{}

	err := k8sClient.Get(
		ctx,
		client.ObjectKey{
			Name:      name,
			Namespace: defaultNamespace,
		},
		pod,
	)
	if err != nil {
		t.Fatalf("get pod %s: %v", name, err)
	}

	actualRole := pod.Labels[defaults.RoleLabel]

	if actualRole != expectedRole {
		t.Fatalf(
			"expected pod %s to have role %q, got %q",
			name,
			expectedRole,
			actualRole,
		)
	}
}
