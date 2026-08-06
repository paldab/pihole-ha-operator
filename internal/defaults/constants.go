// Package defaults contain default values for resource objects
package defaults

import (
	corev1 "k8s.io/api/core/v1"
)

// PiHole cluster
const (
	ApplicationName = "pihole"
	ProbeURL        = "http://localhost/api/info/login"
	DefaultReplicas = 3
	PiholeFTLDBPath = "/etc/pihole/pihole-FTL.db"

	StorageSize         = "3Gi"
	WebserverPort int32 = 80
	DNSPort       int32 = 53
)

var DefaultDNSUpstreams = []string{
	"8.8.8.8",
	"8.8.4.4",
}

var DefaultContainerCapablilties = []corev1.Capability{
	"SYS_NICE",
	"SYS_TIME",
	"NET_BIND_SERVICE",
	"CAP_CHOWN",
	"CAP_NET_BIND_SERVICE",
	"CAP_NET_RAW",
}
