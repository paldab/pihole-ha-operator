// Package builders is a package that build the kubernetes native objects
package builders

import (
	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
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
				WhenDeleted: "Retain", //TODO: This should go to delete whenever we have syncing between all pods
				WhenScaled:  "Retain",
			},
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
					Containers:   containers,
					Affinity:     cluster.Spec.Config.Affinity,
					Tolerations:  cluster.Spec.Config.Tolerations,
					NodeSelector: cluster.Spec.Config.NodeSelector,
				},
			},
		},
	}

	sts.Spec.Template.Spec.Volumes = append(sts.Spec.Template.Spec.Volumes, volumes...)

	return sts
}

func BuildPiholeContainers(cluster piholev1alpha1.PiHoleCluster, volumeMounts []corev1.VolumeMount) []corev1.Container {
	requiredEnvs := defaults.RequiredPiholeEnvs(cluster.Spec.ExistingAdminPasswordSecret, *cluster.Spec.TimeZone, defaults.WebserverPort, cluster.Spec.DNSUpstreams)
	additionalEnvs := defaults.AdditionalPiholeEnvs(&cluster)
	containerEnvs := append(requiredEnvs, additionalEnvs...)
	piholeEnvs := append(containerEnvs, cluster.Spec.Config.Env...)
	containerPorts := defaults.DefaultPiholeContainerPorts(defaults.WebserverPort, defaults.DNSPort, false) // TODO make dynamic dhcp paramter

	return []corev1.Container{
		{
			Name:  defaults.ApplicationName,
			Image: cluster.Spec.Image,
			Ports: containerPorts,
			Env:   piholeEnvs,

			StartupProbe:    cluster.Spec.Config.Probes.Startup,
			ReadinessProbe:  cluster.Spec.Config.Probes.Readiness,
			LivenessProbe:   cluster.Spec.Config.Probes.Liveness,
			SecurityContext: cluster.Spec.Config.SecurityContext,
			VolumeMounts:    volumeMounts,
		},
	}
}
