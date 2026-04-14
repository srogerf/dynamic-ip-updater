package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

type IPLookupClient struct {
	HTTPClient *http.Client
}

var ErrNoPublicIPv6 = errors.New("network does not appear to have a public IPv6 address")

func NewIPLookupClient() *IPLookupClient {
	return &IPLookupClient{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *IPLookupClient) Lookup(ctx context.Context, family, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create %s lookup request: %w", family, err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("perform %s lookup request: %w", family, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 128))
	if err != nil {
		return "", fmt.Errorf("read %s lookup response: %w", family, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%s lookup failed with status %s", family, resp.Status)
	}

	value := strings.TrimSpace(string(body))
	ip := net.ParseIP(value)
	if ip == nil {
		return "", fmt.Errorf("%s lookup returned invalid IP %q", family, value)
	}

	switch family {
	case "ipv4":
		if ip.To4() == nil {
			return "", fmt.Errorf("%s lookup returned non-IPv4 value %q", family, value)
		}
	case "ipv6":
		if ip.To4() != nil {
			return "", fmt.Errorf("%w: lookup service returned IPv4 value %q instead", ErrNoPublicIPv6, value)
		}
	default:
		return "", fmt.Errorf("unsupported IP family %q", family)
	}

	return value, nil
}
