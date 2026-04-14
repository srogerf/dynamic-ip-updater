package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCacheMissingFile(t *testing.T) {
	cache, err := loadCache(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cache.Records) != 0 {
		t.Fatalf("expected empty cache, got %d records", len(cache.Records))
	}
}

func TestSaveAndLoadCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dns-cache.json")
	cache := DNSCache{Records: map[string]string{
		cacheKey("example.com", "@", "A"): "203.0.113.10",
	}}

	if err := saveCache(path, cache); err != nil {
		t.Fatalf("save cache: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat cache: %v", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("unexpected cache mode: %o", info.Mode().Perm())
	}

	loaded, err := loadCache(path)
	if err != nil {
		t.Fatalf("load cache: %v", err)
	}
	if got := getCachedRecord(loaded, "example.com", "@", "A"); got != "203.0.113.10" {
		t.Fatalf("unexpected cached value: %s", got)
	}
}
