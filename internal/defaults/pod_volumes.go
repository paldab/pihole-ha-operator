package defaults

const (
	PiholeStatefulSetVolumeName = "pihole-config"
)

type PiholeVolumeConfig struct {
	Key       string
	MountPath string
}

type PiholeComponent string

const (
	StsVolumeName PiholeComponent = PiholeStatefulSetVolumeName
	Adlist        PiholeComponent = "adlist"
	AddHosts      PiholeComponent = "additional-hosts"
	Custom        PiholeComponent = "custom-config"
	CNAMEs        PiholeComponent = "cname"
)

type PiholeVolumeConfigMap map[PiholeComponent]PiholeVolumeConfig

const (
	VolumeMountAdlistKey   = "adlists.list"
	VolumeMountAddHostsKey = "additional-hosts"
	VolumeMountCustomKey   = "custom.conf"
	VolumeMountCNAMEKey    = "custom-cnames.conf"
)

var PiholeStaticMountConfig = PiholeVolumeConfigMap{
	StsVolumeName: PiholeVolumeConfig{
		Key:       "",
		MountPath: "/etc/pihole",
	},

	Adlist: PiholeVolumeConfig{
		Key:       VolumeMountAdlistKey,
		MountPath: "/etc/pihole/adlists.list",
	},

	AddHosts: PiholeVolumeConfig{
		Key:       VolumeMountAddHostsKey,
		MountPath: "/etc/addn-hosts",
	},

	Custom: PiholeVolumeConfig{
		Key:       VolumeMountCustomKey,
		MountPath: "/etc/dnsmasq.d/02-custom.conf",
	},

	CNAMEs: PiholeVolumeConfig{
		Key:       VolumeMountCNAMEKey,
		MountPath: "/etc/dnsmasq.d/05-custom-cname.conf",
	},
}
