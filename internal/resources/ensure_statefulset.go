package resources

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/internal/builders"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
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

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      defaults.PiholeStatefulSetVolumeName,
			MountPath: "/etc/pihole",
		},
	}

	configVolumes, configVolumeMounts := getPiholeConfigVolumes(rc.Cluster.Name)
	volumeMounts = append(volumeMounts, configVolumeMounts...)
	containers := builders.BuildPiholeContainers(*rc.Cluster, volumeMounts)

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
	volumes := []corev1.Volume{}
	volumeMounts := []corev1.VolumeMount{}

	for component, config := range defaults.PiholeStaticMountConfig {
		configmapName := defaults.GetConfigMapName(clusterName, string(component))
		volumeName := fmt.Sprintf("pihole-%s", component)

		volume := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: configmapName},
				},
			},
		}

		volumeMount := corev1.VolumeMount{
			Name:      volumeName,
			MountPath: config.MountPath,
			SubPath:   config.Key,
		}

		volumes = append(volumes, volume)
		volumeMounts = append(volumeMounts, volumeMount)
	}

	return volumes, volumeMounts
}
