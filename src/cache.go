package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type DNSCache struct {
	Records map[string]string `json:"records"`
}

func loadCache(path string) (DNSCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return DNSCache{Records: map[string]string{}}, nil
		}
		return DNSCache{}, fmt.Errorf("read cache: %w", err)
	}

	var cache DNSCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return DNSCache{}, fmt.Errorf("parse cache: %w", err)
	}
	if cache.Records == nil {
		cache.Records = map[string]string{}
	}

	return cache, nil
}

func saveCache(path string, cache DNSCache) error {
	if cache.Records == nil {
		cache.Records = map[string]string{}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal cache: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}

	return nil
}

func cacheKey(domain, host, recordType string) string {
	return fmt.Sprintf("%s|%s|%s", domain, host, recordType)
}

func getCachedRecord(cache DNSCache, domain, host, recordType string) string {
	return cache.Records[cacheKey(domain, host, recordType)]
}

func setCachedRecord(cache *DNSCache, domain, host, recordType, value string) {
	if cache.Records == nil {
		cache.Records = map[string]string{}
	}
	cache.Records[cacheKey(domain, host, recordType)] = value
}
