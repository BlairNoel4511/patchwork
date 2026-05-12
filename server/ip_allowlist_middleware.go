package server

import (
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/patrickward/patchwork/config"
)

// IPAllowlistMiddleware blocks requests whose remote IP is not in the configured
// allowlist. It supports exact IPs, CIDR ranges, and optional proxy header trust.
func IPAllowlistMiddleware(cfg *config.IPAllowlistConfig, next http.Handler) http.Handler {
	if !cfg.IsEnabled() {
		return next
	}

	// Pre-parse networks once at construction time.
	type entry struct {
		ip  net.IP
		net *net.IPNet
	}
	var entries []entry
	for _, raw := range cfg.Allow {
		if strings.Contains(raw, "/") {
			_, ipNet, err := net.ParseCIDR(raw)
			if err == nil {
				entries = append(entries, entry{net: ipNet})
			}
		} else {
			ip := net.ParseIP(raw)
			if ip != nil {
				entries = append(entries, entry{ip: ip})
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := resolveClientIP(r, cfg.TrustProxy)
		if ip == nil || !isAllowed(ip, entries) {
			http.Error(w, cfg.ResolvedDenyBody(), cfg.ResolvedDenyStatus())
			return
		}
		next.ServeHTTP(w, r)
	})
}

func resolveClientIP(r *http.Request, trustProxy bool) net.IP {
	if trustProxy {
		if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
			// Take the first (leftmost) address.
			parts := strings.SplitN(fwd, ",", 2)
			if ip := net.ParseIP(strings.TrimSpace(parts[0])); ip != nil {
				return ip
			}
		}
		if real := r.Header.Get("X-Real-IP"); real != "" {
			if ip := net.ParseIP(strings.TrimSpace(real)); ip != nil {
				return ip
			}
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// RemoteAddr may already be a bare IP in tests.
		host = r.RemoteAddr
	}
	_ = fmt.Sprintf // suppress unused import warning
	return net.ParseIP(host)
}

type allowEntry struct {
	ip  net.IP
	net *net.IPNet
}

func isAllowed(ip net.IP, entries []struct {
	ip  net.IP
	net *net.IPNet
}) bool {
	for _, e := range entries {
		if e.net != nil && e.net.Contains(ip) {
			return true
		}
		if e.ip != nil && e.ip.Equal(ip) {
			return true
		}
	}
	return false
}
