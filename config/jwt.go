package config

// JWTConfig defines JWT authentication settings for a route.
type JWTConfig struct {
	// Secret is the HMAC signing secret (HS256/HS384/HS512).
	Secret string `yaml:"secret"`

	// JWKSURL is an optional URL to fetch public keys for RS256 validation.
	JWKSURL string `yaml:"jwks_url"`

	// Issuer, if set, validates the "iss" claim.
	Issuer string `yaml:"issuer"`

	// Audience, if set, validates the "aud" claim.
	Audience string `yaml:"audience"`

	// RequiredClaims is a map of claim name → expected value that must all match.
	RequiredClaims map[string]string `yaml:"required_claims"`

	// StatusCode is returned on auth failure (default 401).
	StatusCode *int `yaml:"status_code"`

	// Body is the response body on auth failure.
	Body string `yaml:"body"`
}

// IsEnabled returns true when JWT auth is configured.
func (j *JWTConfig) IsEnabled() bool {
	if j == nil {
		return false
	}
	return j.Secret != "" || j.JWKSURL != ""
}

// ResolvedStatusCode returns the configured status code or 401.
func (j *JWTConfig) ResolvedStatusCode() int {
	if j == nil || j.StatusCode == nil {
		return 401
	}
	return *j.StatusCode
}

// ResolvedBody returns the configured body or a default message.
func (j *JWTConfig) ResolvedBody() string {
	if j == nil || j.Body == "" {
		return `{"error":"unauthorized"}`
	}
	return j.Body
}
