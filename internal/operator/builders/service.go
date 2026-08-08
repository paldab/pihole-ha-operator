package builders

import (
	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func BuildService(cluster *piholev1alpha1.PiHoleCluster, config piholev1alpha1.ServiceConfig, ports []corev1.ServicePort, selectorLabels map[string]string) *corev1.Service {
	svc := &corev1.Service{
		// ObjectMeta is found in the EnsureServices function
		Spec: corev1.ServiceSpec{
			Selector: selectorLabels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}

	if config.Annotations != nil {
		svc.Annotations = config.Annotations
	}

	if config.Type != nil {
		svc.Spec.Type = *config.Type
	}

	if config.LoadBalancerIP != nil {
		svc.Spec.LoadBalancerIP = *config.LoadBalancerIP
	}

	return svc
}
