package main

import (
	"context"
	"errors"
	"log"
	"os"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	if err := run(os.Args[1:]); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cli, err := parseCLI(args)
	if err != nil {
		return err
	}

	exeDir, err := executableDir()
	if err != nil {
		return err
	}

	cli.ConfigPath = resolvePath(exeDir, cli.ConfigPath)
	fileCfg, err := loadConfig(cli.ConfigPath)
	if err != nil {
		return err
	}

	cfg := mergeConfig(fileCfg, cli)
	cfg, err = finalizeConfigPaths(cfg, cli)
	if err != nil {
		return err
	}
	if err := validateConfig(cfg); err != nil {
		return err
	}

	cache, err := loadCache(cfg.CachePath)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ipClient := NewIPLookupClient()
	goDaddy := NewGoDaddyClient(cfg)

	if cfg.EnableIPv4 {
		if err := processRecord(ctx, ipClient, goDaddy, &cache, cfg, recordSpec{
			family:     "ipv4",
			recordType: "A",
			lookupURL:  cfg.IPv4URL,
		}); err != nil {
			return err
		}
	}
	if cfg.EnableIPv6 {
		if err := processRecord(ctx, ipClient, goDaddy, &cache, cfg, recordSpec{
			family:     "ipv6",
			recordType: "AAAA",
			lookupURL:  cfg.IPv6URL,
		}); err != nil {
			if errors.Is(err, ErrNoPublicIPv6) {
				log.Printf("Skipping AAAA update: %v", err)
			} else {
				return err
			}
		}
	}

	return saveCache(cfg.CachePath, cache)
}

type recordSpec struct {
	family     string
	recordType string
	lookupURL  string
}

func processRecord(ctx context.Context, ipClient *IPLookupClient, goDaddy *GoDaddyClient, cache *DNSCache, cfg Config, spec recordSpec) error {
	log.Printf("Looking up current %s address from %s", spec.family, spec.lookupURL)
	currentIP, err := ipClient.Lookup(ctx, spec.family, spec.lookupURL)
	if err != nil {
		return err
	}
	log.Printf("Current %s address: %s", spec.family, currentIP)

	cachedIP := getCachedRecord(*cache, cfg.Domain, cfg.Host, spec.recordType)
	if cachedIP != "" {
		log.Printf("Cached %s DNS address: %s", spec.recordType, cachedIP)
		if cachedIP == currentIP {
			log.Printf("Cached %s record matches current %s address, skipping DNS lookup", spec.recordType, spec.family)
			return nil
		}
	}

	log.Printf("Looking up configured %s DNS record for %s/%s", spec.recordType, cfg.Domain, cfg.Host)
	configuredIP, err := goDaddy.GetRecord(ctx, spec.recordType, cfg.Domain, cfg.Host)
	if err != nil {
		return err
	}
	log.Printf("Configured %s DNS address: %s", spec.recordType, configuredIP)
	setCachedRecord(cache, cfg.Domain, cfg.Host, spec.recordType, configuredIP)

	if configuredIP == currentIP {
		log.Printf("No %s update needed", spec.recordType)
		setCachedRecord(cache, cfg.Domain, cfg.Host, spec.recordType, currentIP)
		return nil
	}

	if cfg.DryRun {
		log.Printf("Dry run: would set %s record for %s/%s to %s", spec.recordType, cfg.Domain, cfg.Host, currentIP)
		return nil
	}

	log.Printf("Setting new %s DNS address to %s", spec.recordType, currentIP)
	if err := goDaddy.SetRecord(ctx, spec.recordType, cfg.Domain, cfg.Host, currentIP); err != nil {
		return err
	}
	log.Printf("Updated %s record for %s/%s", spec.recordType, cfg.Domain, cfg.Host)
	setCachedRecord(cache, cfg.Domain, cfg.Host, spec.recordType, currentIP)

	return nil
}
