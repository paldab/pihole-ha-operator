package defaults

import (
	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func defaultConfigProbes(config *piholev1alpha1.PiHolePodConfig) {
	defaultProbes := DefaultProbesObj()
	if config.Probes == nil {
		config.Probes = &piholev1alpha1.Probes{
			Startup: defaultProbes["Startup"],
		}
	}

	if config.Probes.Readiness == nil {
		config.Probes.Readiness = defaultProbes["Readiness"]
	}

	if config.Probes.Liveness == nil {
		config.Probes.Liveness = defaultProbes["Liveness"]
	}
}

// defaultConfigAffinity builds the preferred pod affinity that prefers not to schedule on a node that already has a pihole pod
func defaultConfigAffinity(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.Config.Affinity != nil {
		return
	}

	clusterLabels := PiholeOperatorLabels(obj.Name)
	obj.Spec.Config.Affinity = &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					Weight: 30,
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: clusterLabels,
						},
						Namespaces: []string{
							obj.Namespace,
						},

						TopologyKey: "kubernetes.io/hostname",
					},
				},
			},
		},
	}
}

func IsDHCPEnabled(services *piholev1alpha1.PiHoleServiceSpec) bool {
	return services != nil &&
		services.DHCP != nil &&
		ptr.Deref(services.DHCP.Enabled, false)
}

func defaultConfigSecurityContext(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.Config.SecurityContext != nil {
		return
	}

	capabilities := DefaultContainerCapablilties

	// Add DHCP capabilities if enabled
	if IsDHCPEnabled(obj.Spec.Services) {
		capabilities = append(capabilities, DHCPContainerCapabilities...)
	}

	obj.Spec.Config.SecurityContext = &corev1.SecurityContext{
		Privileged: new(false),
		Capabilities: &corev1.Capabilities{
			Add:  capabilities,
			Drop: []corev1.Capability{"ALL"},
		},
	}
}
