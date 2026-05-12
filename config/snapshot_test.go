package config

import (
	"testing"
	"time"
)

func boolSnapshotPtr(b bool) *bool { return &b }

func TestSnapshotConfig_IsEnabled_Nil(t *testing.T) {
	var s *SnapshotConfig
	if s.IsEnabled() {
		t.Fatal("nil config should not be enabled")
	}
}

func TestSnapshotConfig_IsEnabled_False(t *testing.T) {
	s := &SnapshotConfig{Enabled: boolSnapshotPtr(false)}
	if s.IsEnabled() {
		t.Fatal("expected disabled")
	}
}

func TestSnapshotConfig_IsEnabled_True(t *testing.T) {
	s := &SnapshotConfig{Enabled: boolSnapshotPtr(true)}
	if !s.IsEnabled() {
		t.Fatal("expected enabled")
	}
}

func TestSnapshotConfig_DirDefault(t *testing.T) {
	s := &SnapshotConfig{}
	if got := s.SnapshotDir(); got != ".patchwork-snapshots" {
		t.Fatalf("unexpected default dir: %s", got)
	}
}

func TestSnapshotConfig_DirCustom(t *testing.T) {
	s := &SnapshotConfig{Dir: "/tmp/snaps"}
	if got := s.SnapshotDir(); got != "/tmp/snaps" {
		t.Fatalf("unexpected dir: %s", got)
	}
}

func TestSnapshotConfig_ParseTTL_Empty(t *testing.T) {
	s := &SnapshotConfig{}
	d, err := s.ParseTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Fatalf("expected zero duration, got %v", d)
	}
}

func TestSnapshotConfig_ParseTTL_Valid(t *testing.T) {
	s := &SnapshotConfig{TTL: "10m"}
	d, err := s.ParseTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 10*time.Minute {
		t.Fatalf("expected 10m, got %v", d)
	}
}

func TestSnapshotConfig_ParseTTL_Invalid(t *testing.T) {
	s := &SnapshotConfig{TTL: "not-a-duration"}
	_, err := s.ParseTTL()
	if err == nil {
		t.Fatal("expected parse error")
	}
}
