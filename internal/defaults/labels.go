package defaults

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
)

const RoleLabel = "paldab.nl/instanceRole"

var PrimaryPodLabels = map[string]string{
	RoleLabel: "primary",
}

var StandbyPodLabels = map[string]string{
	RoleLabel: "standby",
}

func PiholeOperatorLabels(cluster *v1alpha1.PiHoleCluster) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "pihole-ha-operator",
		"paldab.nl/cluster":            cluster.Name,
		// "app.kubernetes.io/version":    imageTag, TODO
	}
}

// Pihole config
const (
	ConfigMapChecksumLabel = "paldab.nl/config-checksum"
)

func GetConfigMapName(clusterName, component string) string {
	return fmt.Sprintf("%s-%s", clusterName, component)
}
