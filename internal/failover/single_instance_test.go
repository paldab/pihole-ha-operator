package failover_test

import (
	"context"
	"testing"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/failover"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	defaultSinglePodName = "pihole-0"
	defaultNamespace     = "default"
)

func TestReconcileFailoverSingleInstance_HealthyPodWithoutRoleBecomesPrimary(
	t *testing.T,
) {
	ctx := context.Background()

	pod := newSingleInstancePod(defaultSinglePodName, "", true)
	k8sClient := newSingleInstanceFakeClient(t, pod.DeepCopy())

	_, err := failover.ReconcileFailoverSingleInstance(
		ctx,
		k8sClient,
		pod.DeepCopy(),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSingleInstancePodRole(
		t,
		ctx,
		k8sClient,
		pod.Namespace,
		pod.Name,
		"primary",
	)
}

func TestReconcileFailoverSingleInstance_HealthyStandbyBecomesPrimary(
	t *testing.T,
) {
	ctx := context.Background()

	pod := newSingleInstancePod(defaultSinglePodName, "standby", true)
	k8sClient := newSingleInstanceFakeClient(t, pod.DeepCopy())

	_, err := failover.ReconcileFailoverSingleInstance(
		ctx,
		k8sClient,
		pod.DeepCopy(),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertSingleInstancePodRole(
		t,
		ctx,
		k8sClient,
		pod.Namespace,
		pod.Name,
		"primary",
	)
}

func TestReconcileFailoverSingleInstance_HealthyPrimaryRemainsPrimary(
	t *testing.T,
) {
	ctx := context.Background()

	pod := newSingleInstancePod(defaultSinglePodName, "primary", true)
	k8sClient := newSingleInstanceFakeClient(t, pod.DeepCopy())

	result, err := failover.ReconcileFailoverSingleInstance(
		ctx,
		k8sClient,
		pod.DeepCopy(),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RequeueAfter != 0 {
		t.Fatalf("expected no requeue, got %+v", result)
	}

	assertSingleInstancePodRole(
		t,
		ctx,
		k8sClient,
		pod.Namespace,
		pod.Name,
		"primary",
	)
}

func TestReconcileFailoverSingleInstance_UnhealthyPodIsNotPromoted(
	t *testing.T,
) {
	ctx := context.Background()

	pod := newSingleInstancePod(defaultSinglePodName, "", false)
	k8sClient := newSingleInstanceFakeClient(t, pod.DeepCopy())

	result, err := failover.ReconcileFailoverSingleInstance(
		ctx,
		k8sClient,
		pod.DeepCopy(),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RequeueAfter <= 0 {
		t.Fatalf(
			"expected unhealthy pod to be requeued, got %+v",
			result,
		)
	}

	assertSingleInstancePodRole(
		t,
		ctx,
		k8sClient,
		pod.Namespace,
		pod.Name,
		"",
	)
}

func TestReconcileFailoverSingleInstance_TerminatingPodIsNotPromoted(
	t *testing.T,
) {
	ctx := context.Background()

	pod := newSingleInstancePod(defaultSinglePodName, "", true)
	now := metav1.Now()
	pod.DeletionTimestamp = &now
	pod.Finalizers = []string{"test.example/finalizer"}

	k8sClient := newSingleInstanceFakeClient(t, pod.DeepCopy())

	result, err := failover.ReconcileFailoverSingleInstance(
		ctx,
		k8sClient,
		pod.DeepCopy(),
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.RequeueAfter <= 0 {
		t.Fatalf(
			"expected terminating pod to be requeued, got %+v",
			result,
		)
	}

	assertSingleInstancePodRole(
		t,
		ctx,
		k8sClient,
		pod.Namespace,
		pod.Name,
		"",
	)
}

func newSingleInstancePod(
	name string, //nolint:unparam
	role string,
	ready bool,
) *corev1.Pod {
	labels := map[string]string{}

	if role != "" {
		labels[defaults.RoleLabel] = role
	}

	readyStatus := corev1.ConditionFalse
	if ready {
		readyStatus = corev1.ConditionTrue
	}

	return &corev1.Pod{
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

func newSingleInstanceFakeClient(
	t *testing.T,
	objects ...client.Object,
) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()

	if err := corev1.AddToScheme(scheme); err != nil {
		t.Fatalf("add core API to scheme: %v", err)
	}

	if err := appsv1.AddToScheme(scheme); err != nil {
		t.Fatalf("add apps API to scheme: %v", err)
	}

	if err := v1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("add PiHole API to scheme: %v", err)
	}

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(objects...).
		Build()
}

func assertSingleInstancePodRole(
	t *testing.T,
	ctx context.Context,
	k8sClient client.Client,
	namespace string,
	name string,
	expectedRole string,
) {
	t.Helper()

	pod := &corev1.Pod{}

	err := k8sClient.Get(
		ctx,
		client.ObjectKey{
			Namespace: namespace,
			Name:      name,
		},
		pod,
	)
	if err != nil {
		t.Fatalf("get pod %s: %v", name, err)
	}

	actualRole := pod.Labels[defaults.RoleLabel]

	if actualRole != expectedRole {
		t.Fatalf(
			"expected pod %s role %q, got %q",
			name,
			expectedRole,
			actualRole,
		)
	}
}
