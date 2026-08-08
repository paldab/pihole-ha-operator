package exporterapi

type PiholeQueryType int

const (
	PiholeQueryTypeA PiholeQueryType = iota + 1
	PiholeQueryTypeAAAA
	PiholeQueryTypeANY
	PiholeQueryTypeSRV
	PiholeQueryTypeSOA
	PiholeQueryTypePTR
	PiholeQueryTypeTXT
	PiholeQueryTypeNAPTR
	PiholeQueryTypeMX
	PiholeQueryTypeDS
	PiholeQueryTypeRRSIG
	PiholeQueryTypeDNSKEY
	PiholeQueryTypeNS
	PiholeQueryTypeSVCB
	PiholeQueryTypeHTTPS
	PiholeQueryTypeOTHER // any query type not covered elsewhere, but see note below
)

func (t PiholeQueryType) String() string {
	if name, ok := PiholeQueryTypeNames[t]; ok {
		return name
	}

	return "OTHER"
}

var PiholeQueryTypeNames = map[PiholeQueryType]string{
	PiholeQueryTypeA:      "A",
	PiholeQueryTypeAAAA:   "AAAA",
	PiholeQueryTypeANY:    "ANY",
	PiholeQueryTypeSRV:    "SRV",
	PiholeQueryTypeSOA:    "SOA",
	PiholeQueryTypePTR:    "PTR",
	PiholeQueryTypeTXT:    "TXT",
	PiholeQueryTypeNAPTR:  "NAPTR",
	PiholeQueryTypeMX:     "MX",
	PiholeQueryTypeDS:     "DS",
	PiholeQueryTypeRRSIG:  "RRSIG",
	PiholeQueryTypeDNSKEY: "DNSKEY",
	PiholeQueryTypeNS:     "NS",
	PiholeQueryTypeSVCB:   "SVCB",
	PiholeQueryTypeHTTPS:  "HTTPS",
	PiholeQueryTypeOTHER:  "OTHER",
}

// PiholeQueryStatus information is found here https://docs.pi-hole.net/database/query-database/#supported-status-types
type PiholeQueryStatus int

const (
	PiholeStatusTypeUnknown PiholeQueryStatus = iota // 0

	// Blocked
	PiholeStatusTypeBlockedGravity              // 1 - Domain in gravity database
	PiholeStatusTypeAllowedForwarded            // 2 - Forwarded
	PiholeStatusTypeAllowedCached               // 3 - Replied from cache
	PiholeStatusTypeBlockedRegex                // 4 - Regex denylist
	PiholeStatusTypeBlockedExact                // 5 - Exact denylist
	PiholeStatusTypeBlockedUpstreamBlockingIP   // 6 - Upstream returned known blocking IP
	PiholeStatusTypeBlockedUpstreamNullAddress  // 7 - Upstream returned 0.0.0.0 or ::
	PiholeStatusTypeBlockedUpstreamNXDomainNoRA // 8 - Upstream returned NXDOMAIN with RA bit unset
	PiholeStatusTypeBlockedCNAMEGravity         // 9 - Gravity during deep CNAME inspection
	PiholeStatusTypeBlockedCNAMERegex           // 10 - Regex during deep CNAME inspection
	PiholeStatusTypeBlockedCNAMEExact           // 11 - Exact denylist during deep CNAME inspection

	// Allowed
	PiholeStatusTypeAllowedRetried          // 12 - Retried query
	PiholeStatusTypeAllowedRetriedIgnored   // 13 - Retried but ignored
	PiholeStatusTypeAllowedAlreadyForwarded // 14 - Already forwarded

	// Blocked
	PiholeStatusTypeBlockedDatabaseBusy  // 15 - Database busy
	PiholeStatusTypeBlockedSpecialDomain // 16 - Special domain
	PiholeStatusTypeAllowedStaleCache    // 17 - Replied from stale cache
	PiholeStatusTypeBlockedUpstreamEDE15 // 18 - Upstream returned EDE 15
)

var PiholeStatusTypeNames = map[PiholeQueryStatus]string{
	PiholeStatusTypeUnknown: "unknown",

	PiholeStatusTypeBlockedGravity:              "blocked_gravity",
	PiholeStatusTypeAllowedForwarded:            "allowed_forwarded",
	PiholeStatusTypeAllowedCached:               "allowed_cached",
	PiholeStatusTypeBlockedRegex:                "blocked_regex",
	PiholeStatusTypeBlockedExact:                "blocked_exact",
	PiholeStatusTypeBlockedUpstreamBlockingIP:   "blocked_upstream_blocking_ip",
	PiholeStatusTypeBlockedUpstreamNullAddress:  "blocked_upstream_null_address",
	PiholeStatusTypeBlockedUpstreamNXDomainNoRA: "blocked_upstream_nxdomain_no_ra",
	PiholeStatusTypeBlockedCNAMEGravity:         "blocked_cname_gravity",
	PiholeStatusTypeBlockedCNAMERegex:           "blocked_cname_regex",
	PiholeStatusTypeBlockedCNAMEExact:           "blocked_cname_exact",

	PiholeStatusTypeAllowedRetried:          "allowed_retried",
	PiholeStatusTypeAllowedRetriedIgnored:   "allowed_retried_ignored",
	PiholeStatusTypeAllowedAlreadyForwarded: "allowed_already_forwarded",

	PiholeStatusTypeBlockedDatabaseBusy:  "blocked_database_busy",
	PiholeStatusTypeBlockedSpecialDomain: "blocked_special_domain",
	PiholeStatusTypeAllowedStaleCache:    "allowed_stale_cache",
	PiholeStatusTypeBlockedUpstreamEDE15: "blocked_upstream_ede15",
}

func (s PiholeQueryStatus) String() string {
	if name, ok := PiholeStatusTypeNames[s]; ok {
		return name
	}

	return "unknown"
}
