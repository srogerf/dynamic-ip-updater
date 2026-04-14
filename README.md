# Dynamic IP Updater

`diu` is a small Go program that finds the current public IP address for this machine and updates DNS records in a GoDaddy-managed account when the value changes.

This version is set up for GoDaddy specifically.

It currently supports:

- IPv4 `A` record updates
- IPv6 `AAAA` record updates
- dry-run mode
- JSON config files
- command-line overrides with strict precedence
- a local DNS cache to avoid unnecessary DNS lookups

## Project Layout

- `src`: Go source code
- `bin`: helper scripts and built binaries
- `etc`: config files and the local DNS cache

## How It Works

For each enabled record type, the updater:

1. Looks up the current public IP from a configured HTTPS endpoint.
2. Validates that the returned value matches the requested IP family.
3. Checks the local cache in `etc/dns-cache.json`.
4. Skips the GoDaddy DNS lookup if the cached value already matches the current public IP.
5. Otherwise reads the current DNS record from GoDaddy.
6. Updates the DNS record only when the current public IP and configured DNS value are different.

IPv4 and IPv6 are handled independently. If the network does not appear to have public IPv6 connectivity, the IPv6 path is skipped with a clear log message instead of failing the whole run.

## Build

Build the standalone binary into `bin`:

```bash
bin/build.sh
```

That creates:

```text
bin/diu
```

## Run

Run the built binary:

```bash
bin/diu
```

Run the Go-based helper script instead:

```bash
bin/diu.sh
```

The script is mainly for local development. The built binary is the better choice when you want to run without a Go toolchain installed on the target machine.

## Config Files

Checked-in example config:

- [etc/config.example.json](/home/roger/vm1/development/company/riffexchange/dynamic-ip-updater/etc/config.example.json:1)

Local real config:

- `etc/config.json`

The real config file is Git-ignored and should contain your actual credentials and DNS target.

Example:

```json
{
  "godaddy_base_url": "https://api.godaddy.com",
  "godaddy_api_key": "YOUR_GODADDY_API_KEY",
  "godaddy_api_secret": "YOUR_GODADDY_API_SECRET",
  "domain": "example.com",
  "host": "@",
  "cache_path": "dns-cache.json",
  "ipv4_url": "https://api.ipify.org",
  "ipv6_url": "https://api64.ipify.org",
  "enable_ipv4": true,
  "enable_ipv6": false,
  "dry_run": true
}
```

Notes:

- Use `@` for the root domain record.
- `host: "@"` means the apex record for `example.com`.
- `cache_path` can be relative to the config file location.
- `ipv6_url` may return IPv4 when the network has no public IPv6. The program detects that and skips the IPv6 update safely.

## Configuration Precedence

The precedence rule is always:

1. command-line values
2. config file values
3. built-in defaults

Example:

```bash
bin/diu --config-file /path/to/config.json --domain riffexchange.com
```

In that case:

- `domain` comes from the command line
- everything else comes from the config file when present
- remaining unset values fall back to defaults

## Command-Line Parameters

All flags support the normal Go forms, for example:

- `--dry-run=true`
- `--dry-run false`
- `--domain riffexchange.com`

Available flags:

- `--config`
  Path to the JSON config file.

- `--config-file`
  Alias for `--config`.

- `--godaddy-base-url`
  GoDaddy API base URL. Defaults to `https://api.godaddy.com`.

- `--godaddy-key`
  GoDaddy API key.

- `--godaddy-secret`
  GoDaddy API secret.

- `--domain`
  Domain to update, for example `riffexchange.com`.

- `--host`
  Host record to update, for example `@` or `www`. Defaults to `@`.

- `--cache`
  Path to the local DNS cache file.

- `--ipv4-url`
  HTTPS endpoint used to look up the current public IPv4 address. Defaults to `https://api.ipify.org`.

- `--ipv6-url`
  HTTPS endpoint used to look up the current public IPv6 address. Defaults to `https://api64.ipify.org`.

- `--dry-run`
  Log what would be updated without changing DNS in GoDaddy.

- `--ipv4-only`
  Only process the IPv4 `A` record.

- `--ipv6-only`
  Only process the IPv6 `AAAA` record.

## Examples

Dry run with the local config:

```bash
bin/diu --dry-run=true
```

Dry run using the helper script:

```bash
bin/diu.sh --dry-run=true
```

Use a different config file:

```bash
bin/diu --config-file /path/to/custom.json
```

Override the domain from the command line:

```bash
bin/diu --config-file /path/to/custom.json --domain riffexchange.com
```

Process only IPv4:

```bash
bin/diu --ipv4-only=true
```

Process only IPv6:

```bash
bin/diu --ipv6-only=true
```

Use alternate lookup URLs:

```bash
bin/diu --ipv4-url https://api.ipify.org --ipv6-url https://api64.ipify.org
```

## Logging

The program logs major steps to standard output, including:

- looking up the current public IP
- looking up the configured DNS record when needed
- setting the new DNS value when an update is required
- skipping an update when the value is already current
- skipping IPv6 safely when no public IPv6 address appears to be available

## Testing

Run the unit tests from `src`:

```bash
GOTOOLCHAIN=local GOCACHE=/tmp/dynamic-ip-updater-gocache go test ./...
```

Or use the same local environment the scripts expect:

```bash
cd src
GOTOOLCHAIN=local GOCACHE=/tmp/dynamic-ip-updater-gocache go test ./...
```

## Security Notes

- Do not commit `etc/config.json`.
- Do not commit `etc/dns-cache.json`.
- Do not commit built binaries from `bin`.
- Prefer config files over command-line secrets when possible.
- The program does not log API keys, secrets, or full authorization headers.
