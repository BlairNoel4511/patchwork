package config

import "time"

// SnapshotConfig controls the response snapshot/recording feature.
type SnapshotConfig struct {
	// Enabled turns snapshot recording on or off.
	Enabled *bool `yaml:"enabled"`

	// Dir is the directory where snapshot files are written.
	// Defaults to ".patchwork-snapshots".
	Dir string `yaml:"dir"`

	// TTL is how long a snapshot is considered fresh before being re-recorded.
	// Zero means snapshots never expire.
	TTL string `yaml:"ttl"`
}

// IsEnabled returns true when snapshot recording is active.
func (s *SnapshotConfig) IsEnabled() bool {
	if s == nil || s.Enabled == nil {
		return false
	}
	return *s.Enabled
}

// SnapshotDir returns the resolved storage directory.
func (s *SnapshotConfig) SnapshotDir() string {
	if s == nil || s.Dir == "" {
		return ".patchwork-snapshots"
	}
	return s.Dir
}

// ParseTTL parses the TTL string and returns the duration.
// Returns (0, nil) when TTL is empty (never expire).
func (s *SnapshotConfig) ParseTTL() (time.Duration, error) {
	if s == nil || s.TTL == "" {
		return 0, nil
	}
	return time.ParseDuration(s.TTL)
}
