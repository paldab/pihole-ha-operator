package resources

import (
	"fmt"
	"slices"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/builders"
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

func buildBaseConfigmap(cluster *v1alpha1.PiHoleCluster, component string) *corev1.ConfigMap {
	configmapName := defaults.GetConfigMapName(cluster.Name, component)
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

func EnsureAdListConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.Adlist))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.Adlists)

		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = v1alpha1.ArrayToString(config.Spec.Adlists)
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountAdlistKey: stringifiedData,
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)

	})

	return err
}

func EnsureCNAMEConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.CNAMEs))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.CNAMEs)
		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = config.Spec.CNAMEs.ToPiholeConfigString()
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountCNAMEKey: stringifiedData,
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)
	})

	return err
}

func EnsureAdditionalHostsConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.AddHosts))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.Hosts)
		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = config.Spec.Hosts.ToPiholeConfigString()
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountAddHostsKey: stringifiedData,
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)
	})

	return err
}

func EnsureCustomConfigMap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.Custom))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.CustomOptions)
		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = config.Spec.CustomOptions.ToPiholeConfigString()
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountCustomKey: stringifiedData,
	}

	desired := builders.BuildConfigmap(configmapStringData)
	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)
	})

	return err
}

func EnsureBlacklistConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.Blacklist))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.Blacklist)

		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = v1alpha1.ArrayToString(config.Spec.Blacklist)
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountBlacklistKey: stringifiedData,
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)
	})

	return err
}

func EnsureWhitelistConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.Whitelist))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.Whitelist)

		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = v1alpha1.ArrayToString(config.Spec.Whitelist)
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountWhitelistKey: stringifiedData,
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

		return ctrl.SetControllerReference(rc.Cluster, configmap, rc.Scheme)
	})

	return err
}

func EnsureRegexListConfigmap(rc *ResourceContext, config *v1alpha1.PiHoleConfig) error {
	configmap := buildBaseConfigmap(rc.Cluster, string(defaults.Regexlist))
	stringifiedData := ""

	if config != nil {
		checksum, err := utils.CalculateChecksum(config.Spec.Regexlist)

		if err != nil {
			return err
		}

		configmap.Annotations = map[string]string{
			checksumAnnotation: checksum,
		}

		stringifiedData = v1alpha1.ArrayToString(config.Spec.Regexlist)
	}

	configmapStringData := map[string]string{
		defaults.VolumeMountRegexlistKey: stringifiedData,
	}

	_, err := controllerutil.CreateOrUpdate(rc.Ctx, rc.K8sClient, configmap, func() error {
		desired := builders.BuildConfigmap(configmapStringData)

		configmap.Data = desired.Data

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
	var error error
	switch component {
	case defaults.Custom:
		error = EnsureCustomConfigMap(rc, nil)
	case defaults.AddHosts:
		error = EnsureAdditionalHostsConfigmap(rc, nil)
	case defaults.CNAMEs:
		error = EnsureCNAMEConfigmap(rc, nil)
	case defaults.Adlist:
		error = EnsureAdListConfigmap(rc, nil)
	case defaults.Blacklist:
		error = EnsureBlacklistConfigmap(rc, nil)
	case defaults.Whitelist:
		error = EnsureWhitelistConfigmap(rc, nil)
	case defaults.Regexlist:
		error = EnsureRegexListConfigmap(rc, nil)
	}

	return error
}
