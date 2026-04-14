# Project Rules

This file captures the agreed rules for the dynamic IP updater before implementation.

## Purpose

- Build a Go program that finds the current public IP address and updates DNS records in a GoDaddy-managed account.

## Rules

### Global Rules

- Never commit secure or identifiable information to Git.
- Never require secrets to be passed on the command line if an environment variable or local config alternative is available.
- Do not commit real domains, API keys, API secrets, public IPs, or other account-identifying values in examples, tests, or docs.

### DNS Behavior

- The updater should target the DNS record configured by `domain`, `host`, and record type.
- The root domain should be represented as `@` in configuration.
- A DNS update should be skipped when the current public IP already matches the cached confirmed DNS value for the same record type.

### GoDaddy Auth And Config

- GoDaddy API authentication must use the official `Authorization: sso-key <API_KEY>:<API_SECRET>` header format.
- The program must support GoDaddy production and OTE environments through configuration.
- The production GoDaddy API base URL should default to `https://api.godaddy.com`.
- The OTE GoDaddy API base URL should default to `https://api.ote-godaddy.com`.
- GoDaddy API keys and secrets must be stored in a local config file or environment variables, never hard-coded in source.
- Any local config file containing keys, secrets, domains, hosts, or other account-identifying values must be Git-ignored and never checked in.
- The program must not log API keys, API secrets, or full authorization headers.
- The initial implementation should target self-serve account usage and must not require reseller-only headers such as `X-Shopper-Id`.

### Public IP Lookup Strategy

- Public IP discovery must use HTTPS-only lookup services.
- The default public IPv4 lookup should be compatible with `https://api.ipify.org`.
- The default public IPv6 lookup should be compatible with `https://api64.ipify.org`.
- Public IPv4 and IPv6 lookup endpoints must be configurable without code changes.
- The configured lookup URL for each IP family must return an address that matches the requested family.
- Lookup responses must be validated as real IP addresses before they can be used for DNS updates.
- IPv4 and IPv6 lookups must be handled independently so one failure does not corrupt the other path.
- IP lookup failures must return clear errors and must never trigger a DNS update with unvalidated data.

### CLI Behavior

- IPv4 updates and IPv6 updates can both be supported, but each must be independently optional through command-line parameters.
- The command-line interface must allow the operator to choose whether to check IPv4, IPv6, or both in a single run.
- The command-line interface must include a dry-run flag.
- The command-line interface must include an IPv4-only flag.
- The command-line interface must include an IPv6-only flag.
- The command-line interface must allow the config file path to be provided directly.
- The command-line interface must allow the lookup URL to be provided directly.
- The command-line interface must allow GoDaddy credentials to be provided directly.
- Configuration resolution must always prefer command-line values first, then the config file, then built-in defaults when available.

### Update Conditions And Safety Checks

- A DNS update must only occur when the validated looked-up public IP address differs from the current DNS record value for the same record type.

### Logging And Error Handling

- Logging should go to standard output for now.
- The program must log each major step of a run.
- The major logged steps must include looking up the current public IP address.
- The major logged steps must include looking up the currently configured DNS IP address.
- The major logged steps must include setting the new DNS IP address when an update is required.
- Logs must remain readable and must not expose secrets or full authorization headers.

### Runtime And Deployment

- Application source code must live under the `src` directory.
- Helper scripts used to run the program must live under the `bin` directory.
- Built binaries in `bin` must be Git-ignored.
- Configuration files must live under the `etc` directory.
- The repository must include an example configuration file that is safe to commit.
- The primary local configuration file used for real runs must not be checked in to Git.

### Secrets Handling

- Secrets and account-identifying values must live in local-only configuration or environment variables.
- Example config files committed to Git must use placeholders only and must never contain real user data.
- Logs, test fixtures, and sample commands must redact or omit secrets.

### Testing Expectations

- The initial implementation should include a simple set of unit tests covering each major part of the code.
- Testing should stay lightweight for the first version and can be expanded later as needed.

## Open Questions

- Which record types should be supported (`A`, `AAAA`, or both)?
- Should updates run once, on an interval, or be designed for cron/systemd?

## Reusable Notes From `../ip_updater`

- Reuse the high-level flow: discover public IP, fetch current DNS record, compare values, update only when needed.
- Keep provider-specific code separate from IP discovery logic.
- Start with a simple CLI-focused implementation instead of reviving the old frontend or daemon server.
- Replace insecure public IP lookup endpoints with secure HTTPS-based providers.
- Replace panic-driven error handling with returned errors and clear logging.
- Add real tests around parsing, comparison, and provider request construction instead of relying on live network calls.
