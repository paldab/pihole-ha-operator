package defaults

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/utils"
)

const (
	RoleLabel    = "paldab.nl/instanceRole"
	LeaderLabel  = "primary"
	StandbyLabel = "standby"
)

var PrimaryPodLabels = map[string]string{
	RoleLabel: LeaderLabel,
}

var StandbyPodLabels = map[string]string{
	RoleLabel: StandbyLabel,
}

func PiholeOperatorLabels(clusterName string) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "pihole",
		"app.kubernetes.io/managed-by": "pihole-ha-operator",
		"app.kubernetes.io/instance":   clusterName,
		"app.kubernetes.io/part-of":    "pihole-ha",

		"paldab.nl/cluster": clusterName,
	}
}

func PiholePodLabels(cluster *v1alpha1.PiHoleCluster) map[string]string {
	staticLabels := StandbyPodLabels
	operatorLabels := PiholeOperatorLabels(cluster.Name)
	userAddedLabels := cluster.Spec.Config.Labels

	operatorEnforcedLabels := utils.MergeMap(operatorLabels, staticLabels)

	// "app.kubernetes.io/version":    imageTag, TODO add only on pods, not on cluster
	if userAddedLabels != nil {
		return utils.MergeMap(userAddedLabels, operatorEnforcedLabels)
	}

	return operatorEnforcedLabels
}

// Pihole config
const (
	ConfigMapChecksumLabel = "paldab.nl/config-checksum"
)

func GetConfigMapName(clusterName, component string) string {
	return fmt.Sprintf("%s-%s", clusterName, component)
}
