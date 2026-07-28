package builders

import (
	corev1 "k8s.io/api/core/v1"
)

// BuildConfigmap builds a configmap with the pihole configuration.
// This function will only build and does not provide logic to filter logic
func BuildConfigmap(data map[string]string) *corev1.ConfigMap {
	configMap := &corev1.ConfigMap{}
	configMap.Data = data

	return configMap
}
