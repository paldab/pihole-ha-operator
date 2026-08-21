package builders

import (
	"fmt"
	"strconv"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	"github.com/paldab/pihole-ha-operator/internal/operator/defaults"
	"github.com/paldab/pihole-ha-operator/version"
	corev1 "k8s.io/api/core/v1"
)

func BuildPiholeContainer(cluster *piholev1alpha1.PiHoleCluster, volumeMounts []corev1.VolumeMount) corev1.Container {
	requiredEnvs := defaults.BasePiholeEnvs(cluster.Spec.ExistingSecretRef, *cluster.Spec.TimeZone, defaults.WebserverPort, cluster.Spec.DNSUpstreams)
	additionalEnvs := defaults.AdditionalPiholeEnvs(cluster)
	containerEnvs := append(requiredEnvs, additionalEnvs...)
	piholeEnvs := append(containerEnvs, cluster.Spec.Config.Env...)

	isDCHPEnabled := defaults.IsDHCPEnabled(cluster.Spec.Services)
	containerPorts := defaults.DefaultPiholeContainerPorts(defaults.WebserverPort, defaults.DNSPort, isDCHPEnabled)

	return corev1.Container{

		Name:  defaults.ApplicationName,
		Image: cluster.Spec.Image,
		Ports: containerPorts,
		Env:   piholeEnvs,

		StartupProbe:    cluster.Spec.Config.Probes.Startup,
		ReadinessProbe:  cluster.Spec.Config.Probes.Readiness,
		LivenessProbe:   cluster.Spec.Config.Probes.Liveness,
		SecurityContext: cluster.Spec.Config.SecurityContext,
		VolumeMounts:    volumeMounts,
	}
}

// BuildStatsExporterContainer assumes StatisticsSyncConfig.External.Mode is other than local
func BuildStatsExporterContainer(clusterUUID string, StatisticsSyncConfig *piholev1alpha1.StatisticsSpec) (corev1.Container, error) {
	if StatisticsSyncConfig.External == nil {
		return corev1.Container{}, fmt.Errorf("statistics are not configured")
	}

	if StatisticsSyncConfig.External.Database.Host == "" {
		return corev1.Container{}, fmt.Errorf("there is no database host configuration detected")
	}

	dbConfig := StatisticsSyncConfig.External.Database

	containerEnvs := []corev1.EnvVar{
		{
			Name:  "CLUSTER_UUID",
			Value: clusterUUID,
		},
		{
			Name:  "SOURCE_ID_DIR",
			Value: defaults.StatisticsExportSourceIDDir,
		},
		{
			Name:  "EXPORTER_BATCH_SIZE",
			Value: strconv.Itoa(int(StatisticsSyncConfig.External.BatchSize)),
		},
		{
			Name:  "EXPORTER_INTERVAL",
			Value: strconv.Itoa(int(StatisticsSyncConfig.External.IntervalSeconds)),
		},
		{
			Name:  "DB_HOST",
			Value: dbConfig.Host,
		},
		{
			Name:  "DB_DATABASE",
			Value: dbConfig.DBName,
		},
		{
			Name: "DB_USER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: dbConfig.SecretRef.UsernameKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: dbConfig.SecretRef.Name,
					},
				},
			},
		},
		{
			Name: "DB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: dbConfig.SecretRef.PasswordKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: dbConfig.SecretRef.Name,
					},
				},
			},
		},
	}

	return corev1.Container{
		Name:  "statistics-exporter",
		Image: fmt.Sprintf("%s:%s", defaults.StatisticsExporterImage, version.Version),
		Env:   containerEnvs,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      string(defaults.StsVolumeName),
				MountPath: "/etc/pihole",
			},
		},
		SecurityContext: &corev1.SecurityContext{
			AllowPrivilegeEscalation: new(false),
			Privileged:               new(false),
			RunAsNonRoot:             new(true),
			RunAsGroup:               new(int64(1000)),
			RunAsUser:                new(int64(1000)),
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			SeccompProfile: &corev1.SeccompProfile{
				Type: corev1.SeccompProfileTypeRuntimeDefault,
			},
		},
	}, nil
}
