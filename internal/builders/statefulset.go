// Package builders is a package that build the kubernetes native objects
package builders

import (
	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/defaults"
	"github.com/paldab/pihole-ha-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BuildPiholeStatefulSet(cluster *piholev1alpha1.PiHoleCluster) *appsv1.StatefulSet {
	clusterLabels := defaults.PiholeOperatorLabels(cluster)
	clusterPodLabels := utils.MergeMap(clusterLabels, cluster.Spec.Config.Labels)
	piholeContainers := buildPiholeContainers(*cluster)

	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
			Labels:    clusterLabels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: cluster.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: clusterLabels,
			},
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: "Retain", // This should go to delete whenever we have syncing between all pods
				WhenScaled:  "Retain",
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pihole-storage",
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
					Labels:      clusterPodLabels,
					Annotations: cluster.Spec.Config.Annotations,
				},

				Spec: corev1.PodSpec{
					Containers:   piholeContainers,
					Affinity:     cluster.Spec.Config.Affinity,
					Tolerations:  cluster.Spec.Config.Tolerations,
					NodeSelector: cluster.Spec.Config.NodeSelector,
				},
			},
		},
	}
}

func buildPiholeContainers(cluster piholev1alpha1.PiHoleCluster) []corev1.Container {
	if cluster.Spec.Config == nil {
		cluster.Spec.Config = &piholev1alpha1.PiHolePodConfig{}
	}

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
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "pihole-storage",
					MountPath: "/etc/pihole",
				},
			},
		},
	}
}
