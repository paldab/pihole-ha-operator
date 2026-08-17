package resources

import (
	"fmt"
	"slices"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/internal/operator/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	checksumAnnotation = "checksum/config"
)

// PiholeConfigItem standarized PiholeConfig data format for configmaps
type PiholeConfigItem interface {
	ToConfigmapString() string
}

func buildBaseConfigmap(cluster *v1alpha1.PiHoleCluster, component defaults.PiholeComponent) *corev1.ConfigMap {
	configmapName := defaults.GetConfigMapName(cluster.Name, string(component))
	operatorLabels := defaults.PiholeOperatorLabels(cluster.Name)

	mergedLabels := utils.MergeMap(operatorLabels, map[string]string{
		"app.kubernetes.io/component": fmt.Sprintf("%s-config", component),
	})

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: cluster.Namespace,
			Labels:    mergedLabels,
		},
	}
}

func EnsureConfigmap(rc *ResourceContext, data PiholeConfigItem, component defaults.PiholeComponent) error {
	configmap := buildBaseConfigmap(rc.Cluster, component)
	stringifiedData := data.ToConfigmapString()
	checksum, err := utils.CalculateChecksum(data)

	if err != nil {
		return err
	}

	desiredData := map[string]string{
		defaults.VolumeMountAdlistKey: stringifiedData,
	}

	_, err = controllerutil.CreateOrPatch(rc.Ctx, rc.K8sClient, configmap, func() error {
		configmap.Data = desiredData

		if checksum != "" {
			if configmap.Annotations == nil {
				configmap.Annotations = map[string]string{}
			}

			configmap.Annotations[checksumAnnotation] = checksum
		}

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)
	})

	return err
}

func CreateInitialEmptyPiholeConfigmaps(rc *ResourceContext) error {
	var existingConfigmaps corev1.ConfigMapList
	operatorLabels := defaults.PiholeOperatorLabels(rc.Cluster.Name)

	err := rc.K8sClient.List(
		rc.Ctx,
		&existingConfigmaps,
		client.MatchingLabels(operatorLabels))

	if err != nil {
		return err
	}

	createdConfigMapNames := make([]string, 0, len(existingConfigmaps.Items))
	for _, cm := range existingConfigmaps.Items {
		createdConfigMapNames = append(createdConfigMapNames, cm.Name)
	}

	for component := range defaults.PiholeStaticMountConfig {
		if component == defaults.StsVolumeName {
			continue
		}

		configmapName := defaults.GetConfigMapName(rc.Cluster.Name, string(component))
		if !slices.Contains(createdConfigMapNames, configmapName) {
			err := CreateConfigmapWrapper(rc, nil, component)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func CreateConfigmapWrapper(rc *ResourceContext, config *v1alpha1.PiHoleConfig, component defaults.PiholeComponent) error {
	var err error

	switch component {
	case defaults.Custom:
		err = EnsureConfigmap(rc, &config.Spec.CustomOptions, component)
	case defaults.AddHosts:
		err = EnsureConfigmap(rc, &config.Spec.Hosts, component)
	case defaults.CNAMEs:
		err = EnsureConfigmap(rc, &config.Spec.CNAMEs, component)
	case defaults.Adlist:
		err = EnsureConfigmap(rc, &config.Spec.Adlists, component)
	case defaults.Denylist:
		err = EnsureConfigmap(rc, &config.Spec.Denylist, component)
	case defaults.Allowlist:
		err = EnsureConfigmap(rc, &config.Spec.Allowlist, component)
	case defaults.Regexlist:
		err = EnsureConfigmap(rc, &config.Spec.Regexlist, component)
	}

	return err
}
