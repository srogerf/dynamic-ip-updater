package main

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestGoDaddyGetRecord(t *testing.T) {
	client := &GoDaddyClient{
		BaseURL:   "https://api.example.test",
		APIKey:    "key",
		APISecret: "secret",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if got := r.Header.Get("Authorization"); got != "sso-key key:secret" {
					t.Fatalf("unexpected auth header: %s", got)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader(`[{"data":"203.0.113.42","name":"www","ttl":600,"type":"A"}]`)),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	value, err := client.GetRecord(context.Background(), "A", "example.com", "www")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "203.0.113.42" {
		t.Fatalf("unexpected record value: %s", value)
	}
}

func TestGoDaddySetRecord(t *testing.T) {
	client := &GoDaddyClient{
		BaseURL:   "https://api.example.test",
		APIKey:    "key",
		APISecret: "secret",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPut {
					t.Fatalf("unexpected method: %s", r.Method)
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read body: %v", err)
				}
				bodyText := string(body)
				if !strings.Contains(bodyText, `"data":"203.0.113.99"`) {
					t.Fatalf("unexpected body: %s", string(body))
				}
				if !strings.Contains(bodyText, `"type":"A"`) {
					t.Fatalf("record type missing from body: %s", bodyText)
				}
				if !strings.Contains(bodyText, `"name":"www"`) {
					t.Fatalf("record name missing from body: %s", bodyText)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Status:     "200 OK",
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			}),
		},
	}

	if err := client.SetRecord(context.Background(), "A", "example.com", "www", "203.0.113.99"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
