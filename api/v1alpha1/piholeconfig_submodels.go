package v1alpha1

import (
	"fmt"
	"net/netip"
	"net/url"
	"strings"

	"github.com/paldab/pihole-ha-operator/internal/operator/utils"
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

// ToConfigmapString maps record to 'cname=example.com,ingress.example.com'
func (records CNAMERecords) ToConfigmapString() string {
	keys := utils.GetSortedKeysFromMap(records)
	stringifiedData := ""

	for _, key := range keys {
		cannonicalNameArray := records[key]

		for _, cannonicalName := range cannonicalNameArray {
			entry := fmt.Sprintf("cname=%s,%s", cannonicalName, key)
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

func (records HostRecords) ToConfigmapString() string {
	domains := utils.GetSortedKeysFromMap(records)
	stringifiedData := ""

	for _, domain := range domains {
		host := records[Domain(domain)]
		stringifiedData = stringifiedData + host.String() + "\t" + domain + "\n"
	}

	return stringifiedData
}

type AdListItem string
type AdList []AdListItem

func (adlist *AdList) ToConfigmapString() string {
	return ArrayToString(*adlist)
}

type ListItem string
type List []ListItem

func (list *List) ToConfigmapString() string {
	return ArrayToString(*list)
}

type RegexListItem string
type RegexList []RegexListItem

func (item RegexListItem) String() string {
	return string(item)
}

func (list *RegexList) ToConfigmapString() string {
	return ArrayToString(*list)
}

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

func ArrayToString[T ~string](array []T) string {
	var b strings.Builder
	for _, item := range array {
		b.WriteString(string(item))
		b.WriteByte('\n')
	}

	return b.String()
}
func (item ListItem) String() string {
	return string(item)
}

type CustomOption string

func (option CustomOption) String() string {
	return string(option)
}

type CustomOptions []CustomOption

func (options *CustomOptions) ToConfigmapString() string {
	stringifiedData := ""
	for _, item := range *options {
		stringifiedData = stringifiedData + item.String() + "\n"
	}

	return stringifiedData
}
