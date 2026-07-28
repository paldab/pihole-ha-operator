package defaults

import (
	"fmt"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
)

func ApplyDefaultConfigValues(obj *piholev1alpha1.PiHoleConfig, loadBalancerIP *string) {
	defaultCustomConfig(obj, loadBalancerIP)
}

func defaultCustomConfig(obj *piholev1alpha1.PiHoleConfig, loadBalancerIP *string) {
	defaultOptions := piholev1alpha1.CustomOptions{
		"addn-hosts=/etc/addn-hosts",
	}
	if obj.Spec.CustomOptions == nil {
		if loadBalancerIP != nil {
			defaultOptions = append(defaultOptions, piholev1alpha1.CustomOption(fmt.Sprintf("dhcp-option=6,%s", *loadBalancerIP)))
		}

		obj.Spec.CustomOptions = defaultOptions
	}
}
