package config

// BasicAuthConfig defines HTTP Basic Authentication settings for a route.
type BasicAuthConfig struct {
	// Users is a map of username -> password pairs.
	Users map[string]string `yaml:"users"`
	// Realm is the authentication realm returned in WWW-Authenticate header.
	Realm string `yaml:"realm"`
	// StatusCode overrides the default 401 response status on failure.
	StatusCode *int `yaml:"status_code"`
	// Body overrides the default response body on failure.
	Body string `yaml:"body"`
}

// IsEnabled returns true when basic auth is configured with at least one user.
func (b *BasicAuthConfig) IsEnabled() bool {
	return b != nil && len(b.Users) > 0
}

// ResolvedRealm returns the configured realm or a sensible default.
func (b *BasicAuthConfig) ResolvedRealm() string {
	if b == nil || b.Realm == "" {
		return "Restricted"
	}
	return b.Realm
}

// ResolvedStatusCode returns the configured status code or 401.
func (b *BasicAuthConfig) ResolvedStatusCode() int {
	if b == nil || b.StatusCode == nil {
		return 401
	}
	return *b.StatusCode
}

// ResolvedBody returns the configured body or a default message.
func (b *BasicAuthConfig) ResolvedBody() string {
	if b == nil || b.Body == "" {
		return "Unauthorized"
	}
	return b.Body
}
