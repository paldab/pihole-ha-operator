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
	Blacklist     PiholeComponent = "blacklist"
	Whitelist     PiholeComponent = "whitelist"
	Regexlist     PiholeComponent = "regexlist"
	Custom        PiholeComponent = "custom-config"
	CNAMEs        PiholeComponent = "cname"
)
const (
	VolumeMountAdlistKey    = "adlists.list"
	VolumeMountAddHostsKey  = "additional-hosts"
	VolumeMountCustomKey    = "custom.conf"
	VolumeMountCNAMEKey     = "custom-cnames.conf"
	VolumeMountBlacklistKey = "blacklist.txt"
	VolumeMountWhitelistKey = "whitelist.txt"
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

	Blacklist: PiholeVolumeConfig{
		Key:       VolumeMountBlacklistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, VolumeMountBlacklistKey),
	},

	Whitelist: PiholeVolumeConfig{
		Key:       VolumeMountWhitelistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, VolumeMountWhitelistKey),
	},

	Regexlist: PiholeVolumeConfig{
		Key:       VolumeMountRegexlistKey,
		MountPath: fmt.Sprintf("%s/%s", PiholeConfigDir, Regexlist),
	},
}
