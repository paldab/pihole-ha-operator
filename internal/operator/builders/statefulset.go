// Package builders is a package that build the kubernetes native objects
package builders

import (
	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildPiholeStatefulSet(cluster *piholev1alpha1.PiHoleCluster, labels, podLabels map[string]string, containers []corev1.Container, volumes []corev1.Volume) *appsv1.StatefulSet {
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: cluster.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: "Delete",
				WhenScaled:  "Delete",
			},
			// this is set because of exporter can run migrations and don't want them to run all at once
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      defaults.PiholeStatefulSetVolumeName,
						Namespace: cluster.Namespace,
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						StorageClassName: cluster.Spec.Storage.StorageClass,
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: *cluster.Spec.Storage.Size,
							},
						},
					},
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.StatefulSetUpdateStrategyType("RollingUpdate"),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: cluster.Spec.Config.Annotations,
				},

				Spec: corev1.PodSpec{
					Containers:                   containers,
					Affinity:                     cluster.Spec.Config.Affinity,
					Tolerations:                  cluster.Spec.Config.Tolerations,
					NodeSelector:                 cluster.Spec.Config.NodeSelector,
					AutomountServiceAccountToken: new(false),
					SecurityContext: &corev1.PodSecurityContext{
						FSGroup:             new(int64(1000)),
						FSGroupChangePolicy: new(corev1.FSGroupChangeOnRootMismatch),

						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
				},
			},
		},
	}

	sts.Spec.Template.Spec.Volumes = append(sts.Spec.Template.Spec.Volumes, volumes...)

	return sts
}
