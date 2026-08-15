package defaults

import (
	"fmt"
	"strconv"
	"strings"

	piholev1alpha1 "github.com/paldab/pihole-ha-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func ApplyDefaultClusterValues(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.ExistingSecretRef.PasswordKey == nil {
		obj.Spec.ExistingSecretRef.PasswordKey = new("password")
	}

	defaultConfig(obj)
	defaultStorage(obj)
	defaultServices(obj)
	defaultIngress(obj)
	defaultDNSUpstream(obj)
	DefaultStatisticsObj(obj)
}

func defaultConfig(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.Config == nil {
		obj.Spec.Config = &piholev1alpha1.PiHolePodConfig{}
	}

	defaultConfigAffinity(obj)
	defaultConfigSecurityContext(obj)
	defaultConfigProbes(obj.Spec.Config)
}

func defaultStorage(obj *piholev1alpha1.PiHoleCluster) {
	defaultQuantity := resource.MustParse(StorageSize)
	if obj.Spec.Storage == nil {
		obj.Spec.Storage = &piholev1alpha1.PiholePodStorage{
			Size: &defaultQuantity,
		}
	}

	if obj.Spec.Storage.Size == nil {
		obj.Spec.Storage.Size = &defaultQuantity
	}
}

func defaultServices(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.Services == nil {
		obj.Spec.Services = &piholev1alpha1.PiHoleServiceSpec{}
	}

	servicesObj := obj.Spec.Services

	// init services
	if servicesObj.Web == nil {
		obj.Spec.Services.Web = &piholev1alpha1.ServiceConfig{}
	}

	if servicesObj.DNS == nil {
		obj.Spec.Services.DNS = &piholev1alpha1.ServiceConfig{}
	}

	if servicesObj.DHCP == nil {
		obj.Spec.Services.DHCP = &piholev1alpha1.ServiceConfig{}
	}

	if servicesObj.Web.Enabled == nil {
		obj.Spec.Services.Web.Enabled = new(true)
	}

	if servicesObj.DNS.Enabled == nil {
		obj.Spec.Services.DNS.Enabled = new(true)
	}

	if servicesObj.DHCP.Enabled == nil {
		obj.Spec.Services.DHCP.Enabled = new(false)
	}
}

func defaultIngress(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.Ingress == nil {
		obj.Spec.Ingress = &piholev1alpha1.PiHoleIngressSpec{
			Enabled: new(false),
		}
	}
}

func defaultDNSUpstream(obj *piholev1alpha1.PiHoleCluster) {
	if obj.Spec.DNSUpstreams == nil {
		obj.Spec.DNSUpstreams = DefaultDNSUpstreams
	}
}

func RequiredPiholeEnvs(secretRef piholev1alpha1.ExistingPasswordSecretRef, timezone string, webserverPort int32, DNSUpstreams []string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "TZ",
			Value: timezone,
		},
		{
			Name:  "FTLCONF_webserver_port",
			Value: strconv.FormatInt(int64(webserverPort), 10),
		},
		{
			Name: "FTLCONF_webserver_api_password",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					Key: *secretRef.PasswordKey,
					LocalObjectReference: corev1.LocalObjectReference{
						Name: secretRef.SecretName,
					},
				},
			},
		},
		{
			Name:  "FTLCONF_files_database",
			Value: PiholeFTLDBPath,
		},
	}
}

func AdditionalPiholeEnvs(cluster *piholev1alpha1.PiHoleCluster) []corev1.EnvVar {
	var envs = []corev1.EnvVar{}

	if cluster.Spec.DNSUpstreams != nil {
		if len(cluster.Spec.DNSUpstreams) > 0 {
			envs = append(envs, corev1.EnvVar{
				Name:  "FTLCONF_dns_upstreams",
				Value: strings.Join(cluster.Spec.DNSUpstreams, ";"),
			})
		}
	}

	if cluster.Spec.Ingress != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "VIRTUAL_HOST",
			Value: *cluster.Spec.Ingress.Host,
		})
	}

	if cluster.Spec.Services.DNS.LoadBalancerIP != nil {
		envs = append(envs, corev1.EnvVar{
			Name:  "ServerIP",
			Value: *cluster.Spec.Services.DNS.LoadBalancerIP,
		})
	}

	return envs
}

func DefaultPiholeContainerPorts(webPort, dnsPort int32, dhcpEnabled bool) []corev1.ContainerPort {
	webPorts := []corev1.ContainerPort{ //nolint:prealloc
		{
			Name:          "http",
			ContainerPort: webPort,
			Protocol:      corev1.Protocol("TCP"),
		},
		{
			Name:          "https",
			ContainerPort: 443,
			Protocol:      corev1.Protocol("TCP"),
		},
	}

	dnsPorts := []corev1.ContainerPort{
		{
			Name:          "dns",
			ContainerPort: dnsPort,
			Protocol:      corev1.Protocol("TCP"),
		},
		{
			Name:          "dns-udp",
			ContainerPort: dnsPort,
			Protocol:      corev1.Protocol("UDP"),
		},
		{
			Name:          "client-udp",
			ContainerPort: 67,
			Protocol:      corev1.Protocol("UDP"),
		},
	}

	dhcpPorts := []corev1.ContainerPort{
		{
			Name:          "client-dhcp",
			ContainerPort: 67,
			Protocol:      corev1.Protocol("TCP"),
		},
	}

	standardContainerPorts := append(webPorts, dnsPorts...)
	if dhcpEnabled {
		standardContainerPorts = append(standardContainerPorts, dhcpPorts...)
	}

	return standardContainerPorts
}

func DefaultProbesObj() map[string]*corev1.Probe {
	var defaultProbeHandler = corev1.ProbeHandler{
		Exec: &corev1.ExecAction{
			Command: []string{
				"/bin/sh",
				"-c",
				fmt.Sprintf("curl --silent %s | jq 'if (.dns | not) then halt_error(1) end'", ProbeURL),
			},
		},
	}

	return map[string]*corev1.Probe{
		"Startup": {
			FailureThreshold: 30,
			TimeoutSeconds:   5,
			PeriodSeconds:    5,
			ProbeHandler:     defaultProbeHandler,
		},
		"Readiness": {
			FailureThreshold:    3,
			InitialDelaySeconds: 5,
			TimeoutSeconds:      3,
			PeriodSeconds:       5,
			ProbeHandler:        defaultProbeHandler,
		},
		"Liveness": {
			FailureThreshold:    5,
			InitialDelaySeconds: 30,
			TimeoutSeconds:      5,
			PeriodSeconds:       5,
			ProbeHandler:        defaultProbeHandler,
		},
	}
}

func DefaultStatisticsObj(obj *piholev1alpha1.PiHoleCluster) {
	var statisticsObj = &piholev1alpha1.StatisticsSpec{}

	clusterStats := obj.Spec.Statistics

	if clusterStats == nil {
		statisticsObj.Mode = piholev1alpha1.StatsModeLocal
		obj.Spec.Statistics = statisticsObj
		return
	}

	if clusterStats.Mode == piholev1alpha1.StatsModeLocal {
		return
	}

	statisticsObj.Mode = piholev1alpha1.StatsModeExternal
	if clusterStats.External == nil {
		statisticsObj.External = &piholev1alpha1.ExternalStatsConfig{
			BatchSize:       StatisticsExportBatchSize,
			IntervalSeconds: StatisticsExportInterval,
		}

		obj.Spec.Statistics = statisticsObj
		return
	}

	if clusterStats.External != nil {
		statisticsObj.External = clusterStats.External

		if clusterStats.External.BatchSize > 0 {
			statisticsObj.External.BatchSize = clusterStats.External.BatchSize
		} else {
			statisticsObj.External.BatchSize = StatisticsExportBatchSize
		}

		if clusterStats.External.IntervalSeconds < 60 {
			clusterStats.External.IntervalSeconds = StatisticsExportInterval
		} else {
			statisticsObj.External.IntervalSeconds = StatisticsExportInterval
		}

		obj.Spec.Statistics = statisticsObj
	}
}
