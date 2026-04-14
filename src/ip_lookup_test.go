package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestIPLookupValidatesIPv4(t *testing.T) {
	client := &IPLookupClient{HTTPClient: &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader("203.0.113.10")),
				Header:     make(http.Header),
			}, nil
		}),
	}}
	ip, err := client.Lookup(context.Background(), "ipv4", "https://lookup.example.test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "203.0.113.10" {
		t.Fatalf("unexpected IP: %s", ip)
	}
}

func TestIPLookupRejectsWrongFamily(t *testing.T) {
	client := &IPLookupClient{HTTPClient: &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader("2001:db8::10")),
				Header:     make(http.Header),
			}, nil
		}),
	}}
	if _, err := client.Lookup(context.Background(), "ipv4", "https://lookup.example.test"); err == nil {
		t.Fatal("expected IPv4 family validation error")
	}
}

func TestIPLookupReportsMissingPublicIPv6Clearly(t *testing.T) {
	client := &IPLookupClient{HTTPClient: &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Status:     "200 OK",
				Body:       io.NopCloser(strings.NewReader("203.0.113.10")),
				Header:     make(http.Header),
			}, nil
		}),
	}}

	_, err := client.Lookup(context.Background(), "ipv6", "https://lookup.example.test")
	if err == nil {
		t.Fatal("expected IPv6 lookup error")
	}
	if !errors.Is(err, ErrNoPublicIPv6) {
		t.Fatalf("expected ErrNoPublicIPv6, got %v", err)
	}
	if !strings.Contains(err.Error(), "network does not appear to have a public IPv6 address") {
		t.Fatalf("unexpected error message: %v", err)
	}
}
