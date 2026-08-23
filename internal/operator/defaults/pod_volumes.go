package defaults

import "fmt"

const (
	PiholeStatefulSetVolumeName = "pihole-config"
	PiholeConfigDir             = "/etc/pihole"
	DNSMasqDir                  = "/etc/dnsmasq.d"
)

type PiholeComponent string

const (
	StsVolumeName PiholeComponent = PiholeStatefulSetVolumeName
	Adlist        PiholeComponent = "adlist"
	AddHosts      PiholeComponent = "additional-hosts"
	Denylist      PiholeComponent = "denylist"
	Allowlist     PiholeComponent = "allowlist"
	Regexlist     PiholeComponent = "regexlist"
	Custom        PiholeComponent = "custom-config"
	CNAMEs        PiholeComponent = "cname"
)
const (
	VolumeMountAdlistKey    = "adlists.list"
	VolumeMountAddHostsKey  = "additional-hosts"
	VolumeMountCustomKey    = "custom.conf"
	VolumeMountCNAMEKey     = "custom-cnames.conf"
	VolumeMountDenylistKey  = "denylist.txt"
	VolumeMountAllowlistKey = "allowlist.txt"
	VolumeMountRegexlistKey = "regex.list"
)

type PiholeVolumeConfig struct {
	Key       string
	MountPath string
}

type PiholeVolumeConfigMap map[PiholeComponent]PiholeVolumeConfig

var PiholeStaticMountConfig = PiholeVolumeConfigMap{
	StsVolumeName: PiholeVolumeConfig{
		Key:       "",
		MountPath: PiholeConfigDir,
	},

	AddHosts: PiholeVolumeConfig{
		Key:       VolumeMountAddHostsKey,
		MountPath: "/etc/addn-hosts",
	},

	Custom: PiholeVolumeConfig{
		Key:       VolumeMountCustomKey,
		MountPath: fmt.Sprintf("%s/%s", DNSMasqDir, VolumeMountCustomKey),
	},

	CNAMEs: PiholeVolumeConfig{
		Key:       VolumeMountCNAMEKey,
		MountPath: fmt.Sprintf("%s/%s", DNSMasqDir, VolumeMountCNAMEKey),
	},

	Adlist: PiholeVolumeConfig{
		Key:       VolumeMountAdlistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, VolumeMountAdlistKey),
	},

	Denylist: PiholeVolumeConfig{
		Key:       VolumeMountDenylistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, VolumeMountDenylistKey),
	},

	Allowlist: PiholeVolumeConfig{
		Key:       VolumeMountAllowlistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, VolumeMountAllowlistKey),
	},

	Regexlist: PiholeVolumeConfig{
		Key:       VolumeMountRegexlistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, VolumeMountRegexlistKey),
	},
}
