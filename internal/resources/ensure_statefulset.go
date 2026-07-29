package resources

import (
	"fmt"

	"github.com/paldab/pihole-ha-operator/internal/builders"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func EnsureStatefulSet(rc *ResourceContext) error {
	operatorLabels := defaults.PiholeOperatorLabels(rc.Cluster)
	podLabels := utils.MergeMap(operatorLabels, rc.Cluster.Spec.Config.Labels)

	totalVolumeMounts := []corev1.VolumeMount{
		{
			Name:      defaults.PiholeStatefulSetVolumeName,
			MountPath: "/etc/pihole",
		},
	}

	configVolumes, configVolumeMounts := getPiholeConfigVolumes(rc.Cluster.Name)
	totalVolumeMounts = append(totalVolumeMounts, configVolumeMounts...)
	piholeContainers := builders.BuildPiholeContainers(*rc.Cluster, totalVolumeMounts)

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rc.Cluster.Name,
			Namespace: rc.Cluster.Namespace,
		},
	}

	_, err := controllerutil.CreateOrPatch(rc.Ctx, rc.K8sClient, sts, func() error {
		isNewObject := sts.CreationTimestamp.IsZero()
		desiredObject := builders.BuildPiholeStatefulSet(rc.Cluster, operatorLabels, podLabels, piholeContainers, configVolumes)

		sts.Labels = desiredObject.Labels
		sts.Annotations = desiredObject.Annotations

		if isNewObject {
			sts.Spec = desiredObject.Spec
		} else {
			// Only update mutable fields
			sts.Spec.Replicas = desiredObject.Spec.Replicas
			sts.Spec.Template = desiredObject.Spec.Template
			sts.Spec.UpdateStrategy = desiredObject.Spec.UpdateStrategy
			sts.Spec.PersistentVolumeClaimRetentionPolicy = desiredObject.Spec.PersistentVolumeClaimRetentionPolicy
		}

		return ctrl.SetControllerReference(rc.Cluster, sts, rc.K8sClient.Scheme())
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
