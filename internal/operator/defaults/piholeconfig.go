package defaults

import (
	"fmt"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
)

func ApplyDefaultConfigValues(obj *piholev1alpha1.PiHoleConfig, cluster *piholev1alpha1.PiHoleCluster) {
	if cluster.Spec.Services.DHCP == nil {
		cluster.Spec.Services.DHCP = &piholev1alpha1.ServiceConfig{}
	}

	defaultCustomConfig(obj, cluster.Spec.Services.DHCP.LoadBalancerIP)
}

func defaultCustomConfig(obj *piholev1alpha1.PiHoleConfig, loadBalancerIP *string) {
	additionalHostsMountPath := fmt.Sprintf("addn-hosts=%s", PiholeStaticMountConfig[VolumeMountAddHostsKey].MountPath)
	defaultOptions := piholev1alpha1.CustomOptions{
		piholev1alpha1.CustomOption(additionalHostsMountPath),
	}

	if obj.Spec.CustomOptions == nil {
		if loadBalancerIP != nil {
			defaultOptions = append(defaultOptions, piholev1alpha1.CustomOption(fmt.Sprintf("dhcp-option=6,%s", *loadBalancerIP)))
		}

		obj.Spec.CustomOptions = defaultOptions
	}
}
