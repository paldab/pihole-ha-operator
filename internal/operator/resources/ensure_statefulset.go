package resources

import (
	"fmt"
	"slices"

	"github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/builders"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func applyStatefulSetMetaData(currentSts, desiredSts *appsv1.StatefulSet) {
	currentSts.Labels = desiredSts.Labels
	currentSts.Annotations = desiredSts.Annotations
}

func applyMutableStatefulSetFields(currentSts, desiredSts *appsv1.StatefulSet) {
	currentSts.Spec.Replicas = desiredSts.Spec.Replicas
	currentSts.Spec.Template = desiredSts.Spec.Template
	currentSts.Spec.UpdateStrategy = desiredSts.Spec.UpdateStrategy
	currentSts.Spec.PersistentVolumeClaimRetentionPolicy = desiredSts.Spec.PersistentVolumeClaimRetentionPolicy
}

func EnsureStatefulSet(rc *ResourceContext) error {
	stsLabels := defaults.PiholeOperatorLabels(rc.Cluster.Name)
	podLabels := defaults.PiholePodLabels(rc.Cluster)
	configVolumes, configVolumeMounts := getPiholeConfigVolumes(rc.Cluster.Name)

	piholeContainer := builders.BuildPiholeContainer(rc.Cluster, configVolumeMounts)
	var containers = []corev1.Container{
		piholeContainer,
	}

	if rc.Cluster.Spec.Statistics.Mode == v1alpha1.StatsModeExternal {
		statsExporterContainer, err := builders.BuildStatsExporterContainer(string(rc.Cluster.UID), rc.Cluster.Spec.Statistics)
		if err == nil {
			containers = append(containers, statsExporterContainer)
		}
	}

	desiredSts := builders.BuildPiholeStatefulSet(rc.Cluster, stsLabels, podLabels, containers, configVolumes)

	currentSts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Cluster.Name,
			Namespace: rc.Cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrPatch(rc.Ctx, rc.K8sClient, currentSts, func() error {
		isNewSts := currentSts.CreationTimestamp.IsZero()

		applyStatefulSetMetaData(currentSts, desiredSts)

		if isNewSts {
			currentSts.Spec = desiredSts.Spec
		} else {
			applyMutableStatefulSetFields(currentSts, desiredSts)
		}

		return ctrl.SetControllerReference(rc.Cluster, currentSts, rc.K8sClient.Scheme())
	})

	return err
}

func getPiholeConfigVolumes(clusterName string) ([]corev1.Volume, []corev1.VolumeMount) {
	maxMounts := len(defaults.PiholeStaticMountConfig)

	// first sort the list. Solving a bug where every reconciliation changes the order of the list causing k8s to think its a new patch
	components := make([]string, 0, maxMounts)
	for component := range defaults.PiholeStaticMountConfig {
		components = append(components, string(component))
	}

	slices.Sort(components)

	volumes := make([]corev1.Volume, 0, maxMounts)
	volumeMounts := make([]corev1.VolumeMount, 0, maxMounts)

	for _, key := range components {
		component := defaults.PiholeComponent(key)
		if component == defaults.StsVolumeName {
			volumeMount := corev1.VolumeMount{
				Name:      string(defaults.StsVolumeName),
				MountPath: "/etc/pihole",
			}

			volumeMounts = append(volumeMounts, volumeMount)
			continue
		}

		config := defaults.PiholeStaticMountConfig[component]
		configmapName := defaults.GetConfigMapName(clusterName, string(component))
		volumeName := fmt.Sprintf("pihole-%s", component)

		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: configmapName},
					Optional:             new(true),
				},
			},
		})

		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			MountPath: config.MountPath,
			SubPath:   config.Key,
		})
	}

	return volumes, volumeMounts
}
