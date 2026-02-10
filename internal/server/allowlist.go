package server

import (
	"context"
	"net"
	"strings"
)

// Allowlist controls which upstream hosts are permitted. It supports three
// entry types: exact hostnames (with subdomain matching), wildcard DNS
// patterns (*.internal), and CIDR ranges (10.0.0.0/8).
//
// A nil Allowlist permits all hosts.
type Allowlist struct {
	cidrs     []*net.IPNet
	wildcards []string // stored as ".suffix" (e.g. ".internal" from "*.internal")
	exact     []string // lowercased hostnames
	resolver  func(ctx context.Context, host string) ([]net.IPAddr, error)
}

// ParseAllowlist parses a comma-separated allowlist string into a structured
// Allowlist. Each entry is classified as:
//   - CIDR if it contains "/" (e.g. "10.0.0.0/8")
//   - Wildcard if it starts with "*." (e.g. "*.internal")
//   - Exact hostname otherwise
//
// Returns nil for an empty string (nil = allow all).
func ParseAllowlist(raw string) *Allowlist {
	if raw == "" {
		return nil
	}

	a := &Allowlist{
		resolver: net.DefaultResolver.LookupIPAddr,
	}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		switch {
		case strings.Contains(entry, "/"):
			_, cidr, err := net.ParseCIDR(entry)
			if err == nil {
				a.cidrs = append(a.cidrs, cidr)
			}
		case strings.HasPrefix(entry, "*."):
			suffix := strings.ToLower(entry[1:]) // keep the dot: ".internal"
			a.wildcards = append(a.wildcards, suffix)
		default:
			a.exact = append(a.exact, strings.ToLower(entry))
		}
	}

	return a
}

// Allows reports whether the given host (which may include a port) is
// permitted by this allowlist. A nil Allowlist permits all hosts.
func (a *Allowlist) Allows(host string) bool {
	if a == nil {
		return true
	}

	// Strip port, lowercase
	hostname := strings.ToLower(host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = strings.ToLower(h)
	}

	// Check exact entries (exact match or subdomain match)
	for _, entry := range a.exact {
		if hostname == entry || strings.HasSuffix(hostname, "."+entry) {
			return true
		}
	}

	// Check wildcard entries (*.internal stored as ".internal")
	for _, suffix := range a.wildcards {
		if strings.HasSuffix(hostname, suffix) && hostname != suffix[1:] {
			return true
		}
	}

	// Check CIDRs
	if len(a.cidrs) > 0 {
		if ip := net.ParseIP(hostname); ip != nil {
			// IP-literal host: check directly
			for _, cidr := range a.cidrs {
				if cidr.Contains(ip) {
					return true
				}
			}
		} else if a.resolver != nil {
			// Hostname: resolve and check each IP
			addrs, err := a.resolver(context.Background(), hostname)
			if err != nil {
				return false // deny on DNS failure
			}
			for _, addr := range addrs {
				for _, cidr := range a.cidrs {
					if cidr.Contains(addr.IP) {
						return true
					}
				}
			}
		}
	}

	return false
}
