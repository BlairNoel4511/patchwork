package config

import (
	"testing"
	"time"
)

func TestCacheConfig_NilSafe(t *testing.T) {
	var c *CacheConfig
	if c.IsEnabled() {
		t.Fatal("nil CacheConfig should not be enabled")
	}
	ttl, err := c.ParseTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl != 60*time.Second {
		t.Fatalf("expected 60s default, got %v", ttl)
	}
}

func TestCacheConfig_NotEnabled(t *testing.T) {
	c := &CacheConfig{Enabled: false, TTL: "10s"}
	if c.IsEnabled() {
		t.Fatal("expected IsEnabled to be false")
	}
}

func TestCacheConfig_ParsesTTL(t *testing.T) {
	c := &CacheConfig{Enabled: true, TTL: "2m"}
	ttl, err := c.ParseTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl != 2*time.Minute {
		t.Fatalf("expected 2m, got %v", ttl)
	}
}

func TestCacheConfig_InvalidTTL(t *testing.T) {
	c := &CacheConfig{Enabled: true, TTL: "not-a-duration"}
	_, err := c.ParseTTL()
	if err == nil {
		t.Fatal("expected error for invalid TTL")
	}
}

func TestCacheConfig_EmptyTTLDefaultsSixty(t *testing.T) {
	c := &CacheConfig{Enabled: true}
	ttl, err := c.ParseTTL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ttl != 60*time.Second {
		t.Fatalf("expected 60s, got %v", ttl)
	}
}
