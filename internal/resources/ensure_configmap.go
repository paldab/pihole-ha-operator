package resources

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/builders"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func buildBaseConfigmap(cluster *v1alpha1.PiHoleCluster, config *v1alpha1.PiHoleConfig, component string) *corev1.ConfigMap {
	configmapName := defaults.GetConfigMapName(cluster.Name, component)
	operatorLabels := defaults.PiholeOperatorLabels(cluster)

	mergedLabels := utils.MergeMap(operatorLabels, map[string]string{
		"app.kubernetes.io/component": fmt.Sprintf("%s-config", component),
	})

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configmapName,
			Namespace: config.Namespace,
			Labels:    mergedLabels,
		},
	}
}

func EnsureAdListConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, config, "adlist")
	checksum, err := utils.CalculateChecksum(config.Spec.Adlists)

	if err != nil {
		return err
	}

	configmap.Annotations = map[string]string{
		"checksum/config": checksum,
	}

	stringifiedData := config.Spec.Adlists.ArrayToString()
	configmapStringData := map[string]string{
		defaults.VolumeMountAdlistKey: stringifiedData,
	}

	_, err = controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(config, configmap, rc.Scheme)
	})

	return err
}

func EnsureCNAMEConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, config, "cname")
	checksum, err := utils.CalculateChecksum(config.Spec.CNAMEs)

	if err != nil {
		return err
	}

	configmap.Annotations = map[string]string{
		"checksum/config": checksum,
	}

	stringifiedData := config.Spec.CNAMEs.ToPiholeConfigString()

	configmapStringData := map[string]string{
		defaults.VolumeMountCNAMEKey: stringifiedData,
	}

	_, err = controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(config, configmap, rc.Scheme)
	})

	return err
}

func EnsureAdditionalHostsConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, config, "additional-hosts")
	checksum, err := utils.CalculateChecksum(config.Spec.Hosts)

	if err != nil {
		return err
	}

	configmap.Annotations = map[string]string{
		"checksum/config": checksum,
	}

	stringifiedData := config.Spec.Hosts.ToPiholeConfigString()

	configmapStringData := map[string]string{
		defaults.VolumeMountAddHostsKey: stringifiedData,
	}

	_, err = controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(config, configmap, rc.Scheme)
	})

	return err
}

func EnsureCustomConfigMap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, config, "custom-config")
	checksum, err := utils.CalculateChecksum(config.Spec.CustomOptions)

	if err != nil {
		return err
	}

	configmap.Annotations = map[string]string{
		"checksum/config": checksum,
	}

	stringifiedData := config.Spec.CustomOptions.ToPiholeConfigString()

	configmapStringData := map[string]string{
		defaults.VolumeMountCustomKey: stringifiedData,
	}

	desired := builders.BuildConfigmap(configmapStringData)
	_, err = controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(config, configmap, rc.Scheme)
	})

	return err
}
