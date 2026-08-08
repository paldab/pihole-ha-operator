package failover

import (
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	corev1 "k8s.io/api/core/v1"
)

func podIsLeader(pod *corev1.Pod) bool {
	return pod.Labels[defaults.RoleLabel] == "primary"
}

func podIsStandby(pod *corev1.Pod) bool {
	return pod.Labels[defaults.RoleLabel] == "standby"
}

// isPodHealthy returns a boolean to verify if the pod meets the condition to be eligible to be a leader
func isPodHealthy(pod *corev1.Pod) bool {
	podIsTerminating := pod.DeletionTimestamp != nil

	if podIsTerminating {
		return false
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}
