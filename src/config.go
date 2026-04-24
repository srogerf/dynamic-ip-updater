package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultConfigPath     = "../etc/config.json"
	defaultCachePath      = "../etc/dns-cache.json"
	defaultGoDaddyBaseURL = "https://api.godaddy.com"
	defaultIPv4URL        = "https://api.ipify.org"
	defaultIPv6URL        = "https://api64.ipify.org"
)

type Config struct {
	GoDaddyBaseURL string `json:"godaddy_base_url"`
	GoDaddyAPIKey  string `json:"godaddy_api_key"`
	GoDaddySecret  string `json:"godaddy_api_secret"`
	Domain         string `json:"domain"`
	Host           string `json:"host"`
	CachePath      string `json:"cache_path"`
	EnableCache    bool   `json:"enable_cache"`
	IPv4URL        string `json:"ipv4_url"`
	IPv6URL        string `json:"ipv6_url"`
	EnableIPv4     bool   `json:"enable_ipv4"`
	EnableIPv6     bool   `json:"enable_ipv6"`
	DryRun         bool   `json:"dry_run"`
}

type CLIOptions struct {
	ConfigPath string
	GoDaddyURL string
	APIKey     string
	APISecret  string
	Domain     string
	Host       string
	CachePath  string
	EnableCache bool
	IPv4URL    string
	IPv6URL    string
	DryRun     bool
	IPv4Only   bool
	IPv6Only   bool
}

type stringFlag struct {
	set   bool
	value string
}

func (s *stringFlag) String() string {
	return s.value
}

func (s *stringFlag) Set(value string) error {
	s.set = true
	s.value = value
	return nil
}

type boolFlag struct {
	set   bool
	value bool
}

func (b *boolFlag) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b *boolFlag) Set(value string) error {
	b.set = true
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "t", "true", "y", "yes":
		b.value = true
	case "0", "f", "false", "n", "no":
		b.value = false
	default:
		return fmt.Errorf("invalid boolean value %q", value)
	}
	return nil
}

func parseCLI(args []string) (CLIOptions, error) {
	fs := flag.NewFlagSet("dynamic-ip-updater", flag.ContinueOnError)

	var configPath stringFlag
	configPath.value = defaultConfigPath
	var apiBaseURL stringFlag
	var apiKey stringFlag
	var apiSecret stringFlag
	var domain stringFlag
	var host stringFlag
	var cachePath stringFlag
	var enableCache boolFlag
	var ipv4URL stringFlag
	var ipv6URL stringFlag
	var dryRun boolFlag
	var ipv4Only boolFlag
	var ipv6Only boolFlag

	fs.Var(&configPath, "config", "path to the JSON config file")
	fs.Var(&configPath, "config-file", "path to the JSON config file")
	fs.Var(&apiBaseURL, "godaddy-base-url", "GoDaddy API base URL")
	fs.Var(&apiKey, "godaddy-key", "GoDaddy API key")
	fs.Var(&apiSecret, "godaddy-secret", "GoDaddy API secret")
	fs.Var(&domain, "domain", "domain to update")
	fs.Var(&host, "host", "host record to update")
	fs.Var(&cachePath, "cache", "path to the DNS cache file")
	fs.Var(&enableCache, "enable-cache", "enable the local DNS cache optimization")
	fs.Var(&ipv4URL, "ipv4-url", "public IPv4 lookup URL")
	fs.Var(&ipv6URL, "ipv6-url", "public IPv6 lookup URL")
	fs.Var(&dryRun, "dry-run", "log intended changes without updating GoDaddy")
	fs.Var(&ipv4Only, "ipv4-only", "only process the IPv4 record")
	fs.Var(&ipv6Only, "ipv6-only", "only process the IPv6 record")

	if err := fs.Parse(args); err != nil {
		return CLIOptions{}, err
	}

	if ipv4Only.value && ipv6Only.value {
		return CLIOptions{}, errors.New("cannot use --ipv4-only and --ipv6-only together")
	}

	return CLIOptions{
		ConfigPath: configPath.value,
		GoDaddyURL: apiBaseURL.value,
		APIKey:     apiKey.value,
		APISecret:  apiSecret.value,
		Domain:     domain.value,
		Host:       host.value,
		CachePath:  cachePath.value,
		EnableCache: enableCache.value,
		IPv4URL:    ipv4URL.value,
		IPv6URL:    ipv6URL.value,
		DryRun:     dryRun.value,
		IPv4Only:   ipv4Only.value,
		IPv6Only:   ipv6Only.value,
	}, nil
}

func executableDir() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve executable path: %w", err)
	}
	return filepath.Dir(exePath), nil
}

func resolvePath(baseDir, path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Clean(filepath.Join(baseDir, path))
}

func loadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

func mergeConfig(fileCfg Config, cli CLIOptions) Config {
	cfg := Config{
		GoDaddyBaseURL: defaultGoDaddyBaseURL,
		Host:           "@",
		CachePath:      defaultCachePath,
		IPv4URL:        defaultIPv4URL,
		IPv6URL:        defaultIPv6URL,
	}

	if fileCfg.GoDaddyBaseURL != "" {
		cfg.GoDaddyBaseURL = fileCfg.GoDaddyBaseURL
	}
	if fileCfg.GoDaddyAPIKey != "" {
		cfg.GoDaddyAPIKey = fileCfg.GoDaddyAPIKey
	}
	if fileCfg.GoDaddySecret != "" {
		cfg.GoDaddySecret = fileCfg.GoDaddySecret
	}
	if fileCfg.Domain != "" {
		cfg.Domain = fileCfg.Domain
	}
	if fileCfg.Host != "" {
		cfg.Host = fileCfg.Host
	}
	if fileCfg.CachePath != "" {
		cfg.CachePath = fileCfg.CachePath
	}
	cfg.EnableCache = fileCfg.EnableCache
	if fileCfg.IPv4URL != "" {
		cfg.IPv4URL = fileCfg.IPv4URL
	}
	if fileCfg.IPv6URL != "" {
		cfg.IPv6URL = fileCfg.IPv6URL
	}
	cfg.EnableIPv4 = fileCfg.EnableIPv4
	cfg.EnableIPv6 = fileCfg.EnableIPv6
	cfg.DryRun = fileCfg.DryRun

	if cli.GoDaddyURL != "" {
		cfg.GoDaddyBaseURL = cli.GoDaddyURL
	}
	if cli.APIKey != "" {
		cfg.GoDaddyAPIKey = cli.APIKey
	}
	if cli.APISecret != "" {
		cfg.GoDaddySecret = cli.APISecret
	}
	if cli.Domain != "" {
		cfg.Domain = cli.Domain
	}
	if cli.Host != "" {
		cfg.Host = cli.Host
	}
	if cli.CachePath != "" {
		cfg.CachePath = cli.CachePath
	}
	if cli.EnableCache {
		cfg.EnableCache = true
	}
	if cli.IPv4URL != "" {
		cfg.IPv4URL = cli.IPv4URL
	}
	if cli.IPv6URL != "" {
		cfg.IPv6URL = cli.IPv6URL
	}
	if cli.DryRun {
		cfg.DryRun = true
	}

	if cli.IPv4Only {
		cfg.EnableIPv4 = true
		cfg.EnableIPv6 = false
	}
	if cli.IPv6Only {
		cfg.EnableIPv4 = false
		cfg.EnableIPv6 = true
	}
	if !cli.IPv4Only && !cli.IPv6Only && !fileCfg.EnableIPv4 && !fileCfg.EnableIPv6 {
		cfg.EnableIPv4 = true
	}

	return cfg
}

func finalizeConfigPaths(cfg Config, cli CLIOptions) (Config, error) {
	exeDir, err := executableDir()
	if err != nil {
		return Config{}, err
	}

	// Default paths are resolved from the executable location so the built binary
	// can find ../etc/config.json and ../etc/dns-cache.json when launched from bin.
	configPath := resolvePath(exeDir, cli.ConfigPath)
	configDir := filepath.Dir(configPath)

	// A custom cache path in the config file is resolved relative to that config file.
	if cfg.CachePath == defaultCachePath || cfg.CachePath == "" {
		cfg.CachePath = resolvePath(exeDir, cfg.CachePath)
	} else {
		cfg.CachePath = resolvePath(configDir, cfg.CachePath)
	}

	return cfg, nil
}

func validateConfig(cfg Config) error {
	if cfg.Domain == "" {
		return errors.New("domain is required")
	}
	if cfg.GoDaddyAPIKey == "" {
		return errors.New("GoDaddy API key is required")
	}
	if cfg.GoDaddySecret == "" {
		return errors.New("GoDaddy API secret is required")
	}
	if !cfg.EnableIPv4 && !cfg.EnableIPv6 {
		return errors.New("at least one of IPv4 or IPv6 must be enabled")
	}
	if cfg.EnableIPv4 && cfg.IPv4URL == "" {
		return errors.New("IPv4 lookup URL is required when IPv4 is enabled")
	}
	if cfg.EnableIPv6 && cfg.IPv6URL == "" {
		return errors.New("IPv6 lookup URL is required when IPv6 is enabled")
	}
	return nil
}
