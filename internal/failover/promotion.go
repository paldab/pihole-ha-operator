package failover

import (
	"context"

	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// promoteToLeader promotes a pod to leader by changing the role label to primary
func promoteToLeader(ctx context.Context, k8sClient client.Client, newPrimaryPod *corev1.Pod) error {
	originalPod := newPrimaryPod.DeepCopy()

	if newPrimaryPod.Labels == nil {
		newPrimaryPod.Labels = make(map[string]string)
	}

	newPrimaryPod.Labels = utils.MergeMap(newPrimaryPod.Labels, defaults.PrimaryPodLabels)

	return k8sClient.Patch(ctx, newPrimaryPod, client.MergeFrom(originalPod))
}

// demoteToStandby demotes a pod from leader by changing the role label to standy
func demoteToStandby(ctx context.Context, k8sClient client.Client, pod *corev1.Pod) error {
	originalPod := pod.DeepCopy()

	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	pod.Labels = utils.MergeMap(pod.Labels, defaults.StandbyPodLabels)

	return k8sClient.Patch(ctx, pod, client.MergeFrom(originalPod))
}
