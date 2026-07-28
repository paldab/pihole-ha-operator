package v1alpha1

import (
	"fmt"
	"net/netip"
	"net/url"
)

type ClusterRef struct {
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`
}

type CanonicalName string

// CNAMERecords describes the CNAME object for pihole local DNS
// key = CNAMEAlias
// value = CanonicalName
// Example
// traefik.local.home.paldab.nl: ['photos.home.paldab.nl']
// traefik.local.home.paldab.nl:
//   - photos.home.paldab.nl
//   - home.paldab.nl
type CNAMERecords map[string][]CanonicalName

// ToPiholeConfigString maps record to 'cname=example.com,ingress.example.com'
func (records CNAMERecords) ToPiholeConfigString() string {
	stringifiedData := ""

	for destKey, cannonicalNameArray := range records {
		for _, cannonicalName := range cannonicalNameArray {
			entry := fmt.Sprintf("cname=%s,%s", cannonicalName, destKey)
			stringifiedData = stringifiedData + entry + "\n"
		}
	}

	return stringifiedData
}

type Domain string

func (domain Domain) String() string {
	return string(domain)
}

type IPAddress string

func (ip IPAddress) String() string {
	return string(ip)
}

func (ip IPAddress) IsValid() bool {
	addr, err := netip.ParseAddr(string(ip))
	return err == nil && addr.Is4()
}

type HostRecords map[Domain]IPAddress

func (records *HostRecords) ToPiholeConfigString() string {
	stringifiedData := ""

	for domain, host := range *records {
		// entry := fmt.Sprintf("%s%s", host, domain)
		stringifiedData = stringifiedData + host.String() + "\t" + domain.String() + "\n"
	}

	return stringifiedData
}

type AdListItem string
type AdList []AdListItem

func (item AdListItem) String() string {
	return string(item)
}

func (item AdListItem) IsValid() bool {
	u, err := url.ParseRequestURI(string(item))
	if err != nil {
		return false
	}

	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

func (adlist *AdList) ArrayToString() string {
	stringifiedData := ""
	for _, item := range *adlist {
		stringifiedData = stringifiedData + item.String() + "\n"
	}

	return stringifiedData
}

type CustomOption string

func (option CustomOption) String() string {
	return string(option)
}

type CustomOptions []CustomOption

func (options *CustomOptions) ToPiholeConfigString() string {
	stringifiedData := ""
	for _, item := range *options {
		stringifiedData = stringifiedData + item.String() + "\n"
	}

	return stringifiedData
}
