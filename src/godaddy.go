package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GoDaddyClient struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	HTTPClient *http.Client
}

type dnsRecord struct {
	Data string `json:"data"`
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
	Type string `json:"type"`
}

func NewGoDaddyClient(cfg Config) *GoDaddyClient {
	return &GoDaddyClient{
		BaseURL:   strings.TrimRight(cfg.GoDaddyBaseURL, "/"),
		APIKey:    cfg.GoDaddyAPIKey,
		APISecret: cfg.GoDaddySecret,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *GoDaddyClient) GetRecord(ctx context.Context, recordType, domain, host string) (string, error) {
	req, err := c.newRequest(ctx, http.MethodGet, fmt.Sprintf("%s/v1/domains/%s/records/%s/%s", c.BaseURL, domain, recordType, host), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("get %s record: %w", recordType, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return "", fmt.Errorf("read %s record response: %w", recordType, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("get %s record failed with status %s: %s", recordType, resp.Status, strings.TrimSpace(string(body)))
	}

	var records []dnsRecord
	if err := json.Unmarshal(body, &records); err != nil {
		return "", fmt.Errorf("parse %s record response: %w", recordType, err)
	}
	if len(records) == 0 {
		return "", fmt.Errorf("no %s record found for %s/%s", recordType, domain, host)
	}

	return records[0].Data, nil
}

func (c *GoDaddyClient) SetRecord(ctx context.Context, recordType, domain, host, ip string) error {
	payload := []dnsRecord{{Data: ip, TTL: 600}}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal %s record update: %w", recordType, err)
	}

	req, err := c.newRequest(ctx, http.MethodPut, fmt.Sprintf("%s/v1/domains/%s/records/%s/%s", c.BaseURL, domain, recordType, host), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("set %s record: %w", recordType, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return fmt.Errorf("read %s update response: %w", recordType, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("set %s record failed with status %s: %s", recordType, resp.Status, strings.TrimSpace(string(respBody)))
	}

	return nil
}

func (c *GoDaddyClient) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create GoDaddy request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "sso-key "+c.APIKey+":"+c.APISecret)
	return req, nil
}
