// Package safehttp provides HTTP clients that refuse to connect to private,
// loopback and link-local addresses. It is used by the crawler, link checker
// and PDF downloader to prevent SSRF.
package safehttp

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// ErrPrivateAddress is returned when a target resolves to a non-public IP.
	ErrPrivateAddress = errors.New("target resolves to a private or reserved address")
	// ErrUnsupportedScheme is returned for non-HTTP(S) URLs.
	ErrUnsupportedScheme = errors.New("only http and https URLs are allowed")
)

// Resolver resolves a hostname to IP addresses. Tests can replace it.
type Resolver func(ctx context.Context, host string) ([]net.IP, error)

// Policy decides which hosts may be scanned and dials only public addresses.
type Policy struct {
	// AllowedHosts is an explicit allowlist for local development and demos,
	// for example "demo" when the demo site runs in the same Docker network.
	AllowedHosts map[string]struct{}
	// Resolve is the hostname resolver. Defaults to net.DefaultResolver.
	Resolve Resolver
}

// NewPolicy creates a policy with an optional hostname allowlist.
func NewPolicy(allowedHosts []string) *Policy {
	allowed := make(map[string]struct{}, len(allowedHosts))
	for _, host := range allowedHosts {
		host = normalizeHost(host)
		if host != "" {
			allowed[host] = struct{}{}
		}
	}

	return &Policy{
		AllowedHosts: allowed,
		Resolve:      defaultResolver,
	}
}

// AllowsHost reports whether the host is explicitly allowlisted.
func (p *Policy) AllowsHost(host string) bool {
	if p == nil {
		return false
	}
	_, ok := p.AllowedHosts[normalizeHost(host)]
	return ok
}

// ValidateURL checks the scheme and resolves the host. It rejects private,
// loopback, link-local and other reserved addresses unless the host is
// explicitly allowlisted.
func (p *Policy) ValidateURL(ctx context.Context, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ErrUnsupportedScheme
	}

	return p.ValidateHost(ctx, parsed.Hostname())
}

// ValidateHost resolves a hostname and rejects it if any address is not public.
func (p *Policy) ValidateHost(ctx context.Context, host string) error {
	if p == nil {
		return nil
	}
	if p.AllowsHost(host) {
		return nil
	}

	ips, err := p.resolve(ctx, host)
	if err != nil {
		return fmt.Errorf("resolve %s: %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("resolve %s: no addresses", host)
	}

	for _, ip := range ips {
		if !IsPublic(ip) {
			return ErrPrivateAddress
		}
	}

	return nil
}

// DialContext resolves the target itself and dials only a verified public IP.
// This also protects against DNS rebinding: the address that was checked is the
// address that is dialed.
func (p *Policy) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if p != nil && p.AllowsHost(host) {
		var dialer net.Dialer
		return dialer.DialContext(ctx, network, address)
	}

	ips, err := p.resolve(ctx, host)
	if err != nil {
		return nil, fmt.Errorf("resolve %s: %w", host, err)
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("resolve %s: no addresses", host)
	}

	for _, ip := range ips {
		if !IsPublic(ip) {
			return nil, ErrPrivateAddress
		}
	}

	var lastErr error
	for _, ip := range ips {
		dialer := net.Dialer{Timeout: 10 * time.Second}
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}

	return nil, lastErr
}

func (p *Policy) resolve(ctx context.Context, host string) ([]net.IP, error) {
	resolver := defaultResolver
	if p != nil && p.Resolve != nil {
		resolver = p.Resolve
	}
	return resolver(ctx, host)
}

// NewClient creates an HTTP client with a safe transport.
func NewClient(policy *Policy, timeout time.Duration) *http.Client {
	transport := &http.Transport{
		Proxy:                 nil, // a proxy would bypass the dial-time checks
		DialContext:           policy.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          32,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	return &http.Client{Timeout: timeout, Transport: transport}
}

// IsPublic reports whether an IP is a global unicast address suitable for
// scanning. Loopback, private, link-local, multicast and CGNAT ranges are not.
func IsPublic(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return false
	}

	if ip4 := ip.To4(); ip4 != nil {
		// 0.0.0.0/8
		if ip4[0] == 0 {
			return false
		}
		// 100.64.0.0/10 (CGNAT)
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
		// 192.0.0.0/24 (IETF protocol assignments)
		if ip4[0] == 192 && ip4[1] == 0 && ip4[2] == 0 {
			return false
		}
		// 198.18.0.0/15 (benchmarking)
		if ip4[0] == 198 && (ip4[1] == 18 || ip4[1] == 19) {
			return false
		}
	}

	return true
}

func defaultResolver(ctx context.Context, host string) ([]net.IP, error) {
	addrs, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.IP)
	}
	return ips, nil
}

func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}
