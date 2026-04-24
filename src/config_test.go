package main

import (
	"path/filepath"
	"testing"
)

func TestParseCLIConfigFileAlias(t *testing.T) {
	cli, err := parseCLI([]string{"--config-file", "/tmp/custom.json"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cli.ConfigPath != "/tmp/custom.json" {
		t.Fatalf("unexpected config path: %s", cli.ConfigPath)
	}
}

func TestMergeConfigDefaultsToIPv4(t *testing.T) {
	cfg := mergeConfig(Config{}, CLIOptions{})

	if !cfg.EnableIPv4 {
		t.Fatal("expected IPv4 to default to enabled")
	}
	if cfg.EnableIPv6 {
		t.Fatal("expected IPv6 to default to disabled")
	}
	if cfg.GoDaddyBaseURL != defaultGoDaddyBaseURL {
		t.Fatalf("unexpected default GoDaddy base URL: %s", cfg.GoDaddyBaseURL)
	}
	if cfg.CachePath != defaultCachePath {
		t.Fatalf("unexpected default cache path: %s", cfg.CachePath)
	}
	if cfg.EnableCache {
		t.Fatal("expected cache to default to disabled")
	}
}

func TestMergeConfigCLIOverridesFile(t *testing.T) {
	fileCfg := Config{
		Domain:     "example.com",
		Host:       "@",
		EnableIPv4: true,
	}
	cli := CLIOptions{
		Domain:   "override.example.com",
		Host:     "www",
		IPv6Only: true,
	}

	cfg := mergeConfig(fileCfg, cli)

	if cfg.Domain != "override.example.com" {
		t.Fatalf("expected CLI domain override, got %s", cfg.Domain)
	}
	if cfg.Host != "www" {
		t.Fatalf("expected CLI host override, got %s", cfg.Host)
	}
	if cfg.EnableIPv4 {
		t.Fatal("expected IPv4 disabled by --ipv6-only")
	}
	if !cfg.EnableIPv6 {
		t.Fatal("expected IPv6 enabled by --ipv6-only")
	}
}

func TestMergeConfigCLIAlwaysTakesPrecedence(t *testing.T) {
	fileCfg := Config{
		Domain:         "from-config.example.com",
		Host:           "@",
		GoDaddyBaseURL: "https://config.example.test",
		CachePath:      "../etc/from-config-cache.json",
		EnableCache:    false,
		IPv4URL:        "https://ipv4-config.example.test",
		IPv6URL:        "https://ipv6-config.example.test",
		EnableIPv4:     true,
		EnableIPv6:     false,
	}
	cli := CLIOptions{
		Domain:     "from-cli.example.com",
		Host:       "www",
		GoDaddyURL: "https://cli.example.test",
		CachePath:  "/tmp/from-cli-cache.json",
		EnableCache: true,
		IPv4URL:    "https://ipv4-cli.example.test",
		IPv6URL:    "https://ipv6-cli.example.test",
		IPv6Only:   true,
	}

	cfg := mergeConfig(fileCfg, cli)

	if cfg.Domain != "from-cli.example.com" {
		t.Fatalf("expected CLI domain precedence, got %s", cfg.Domain)
	}
	if cfg.Host != "www" {
		t.Fatalf("expected CLI host precedence, got %s", cfg.Host)
	}
	if cfg.GoDaddyBaseURL != "https://cli.example.test" {
		t.Fatalf("expected CLI base URL precedence, got %s", cfg.GoDaddyBaseURL)
	}
	if cfg.CachePath != "/tmp/from-cli-cache.json" {
		t.Fatalf("expected CLI cache precedence, got %s", cfg.CachePath)
	}
	if !cfg.EnableCache {
		t.Fatal("expected CLI enable-cache precedence")
	}
	if cfg.IPv4URL != "https://ipv4-cli.example.test" {
		t.Fatalf("expected CLI IPv4 URL precedence, got %s", cfg.IPv4URL)
	}
	if cfg.IPv6URL != "https://ipv6-cli.example.test" {
		t.Fatalf("expected CLI IPv6 URL precedence, got %s", cfg.IPv6URL)
	}
	if cfg.EnableIPv4 {
		t.Fatal("expected CLI IPv6-only precedence to disable IPv4")
	}
	if !cfg.EnableIPv6 {
		t.Fatal("expected CLI IPv6-only precedence to enable IPv6")
	}
}

func TestParseCLIEnableCache(t *testing.T) {
	cli, err := parseCLI([]string{"--enable-cache", "true"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cli.EnableCache {
		t.Fatal("expected enable-cache to parse as true")
	}
}

func TestValidateConfigRequiresSecretsAndDomain(t *testing.T) {
	err := validateConfig(Config{EnableIPv4: true, IPv4URL: defaultIPv4URL})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestResolvePath(t *testing.T) {
	got := resolvePath("/tmp/bin", "../etc/config.json")
	want := filepath.Clean("/tmp/etc/config.json")
	if got != want {
		t.Fatalf("unexpected resolved path: got %s want %s", got, want)
	}
}
