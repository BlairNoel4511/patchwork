package config

import "strings"

// IPAllowlistConfig restricts route access to specific IP ranges or addresses.
type IPAllowlistConfig struct {
	// Allow is a list of allowed CIDRs or exact IPs (e.g. "192.168.1.0/24", "10.0.0.1").
	Allow []string `yaml:"allow"`
	// DenyStatus is the HTTP status code returned when a request is denied (default: 403).
	DenyStatus *int `yaml:"deny_status"`
	// DenyBody is the response body returned when a request is denied.
	DenyBody string `yaml:"deny_body"`
	// TrustProxy controls whether X-Forwarded-For / X-Real-IP headers are trusted.
	TrustProxy bool `yaml:"trust_proxy"`
}

// IsEnabled returns true when the allowlist has at least one entry.
func (c *IPAllowlistConfig) IsEnabled() bool {
	return c != nil && len(c.Allow) > 0
}

// ResolvedDenyStatus returns the configured deny status or 403.
func (c *IPAllowlistConfig) ResolvedDenyStatus() int {
	if c == nil || c.DenyStatus == nil {
		return 403
	}
	return *c.DenyStatus
}

// ResolvedDenyBody returns the configured deny body or a default message.
func (c *IPAllowlistConfig) ResolvedDenyBody() string {
	if c == nil || strings.TrimSpace(c.DenyBody) == "" {
		return "forbidden"
	}
	return c.DenyBody
}
