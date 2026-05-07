package config

// ProxyConfig defines upstream forwarding behaviour for a route.
// When set, patchwork will forward the incoming request to the Target URL
// and relay the upstream response back to the caller instead of serving a
// static mock response.
type ProxyConfig struct {
	// Target is the base URL of the upstream service, e.g. "https://api.example.com".
	Target string `yaml:"target"`

	// PassThrough, when true, still executes the normal mock response pipeline
	// after the proxy call completes (useful for logging / chaos injection).
	// Defaults to false — the proxy response is returned directly.
	PassThrough bool `yaml:"pass_through"`
}

// IsEnabled returns true when a non-empty target has been configured.
func (p *ProxyConfig) IsEnabled() bool {
	return p != nil && p.Target != ""
}
