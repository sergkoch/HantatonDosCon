package safehttp

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"testing"
)

func fakeResolver(ips ...string) Resolver {
	return func(context.Context, string) ([]net.IP, error) {
		result := make([]net.IP, 0, len(ips))
		for _, raw := range ips {
			result = append(result, net.ParseIP(raw))
		}
		return result, nil
	}
}

func TestIsPublic(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"8.8.8.8", true},
		{"93.184.216.34", true},
		{"127.0.0.1", false},
		{"10.0.0.1", false},
		{"172.16.0.1", false},
		{"192.168.1.1", false},
		{"169.254.169.254", false},
		{"100.64.0.1", false},
		{"0.0.0.0", false},
		{"198.18.0.1", false},
		{"::1", false},
		{"fc00::1", false},
		{"fe80::1", false},
	}

	for _, test := range tests {
		if got := IsPublic(net.ParseIP(test.ip)); got != test.want {
			t.Fatalf("IsPublic(%s) = %v, want %v", test.ip, got, test.want)
		}
	}
}

func TestValidateURL(t *testing.T) {
	policy := NewPolicy(nil)
	policy.Resolve = fakeResolver("127.0.0.1")

	if err := policy.ValidateURL(context.Background(), "http://private.test/"); !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("expected ErrPrivateAddress, got %v", err)
	}

	policy.Resolve = fakeResolver("93.184.216.34")
	if err := policy.ValidateURL(context.Background(), "https://public.test/page"); err != nil {
		t.Fatalf("expected public URL to pass, got %v", err)
	}

	if err := policy.ValidateURL(context.Background(), "ftp://public.test/file"); !errors.Is(err, ErrUnsupportedScheme) {
		t.Fatalf("expected ErrUnsupportedScheme, got %v", err)
	}
}

func TestAllowlist(t *testing.T) {
	policy := NewPolicy([]string{"demo"})
	policy.Resolve = fakeResolver("172.18.0.2")

	if err := policy.ValidateURL(context.Background(), "http://demo/"); err != nil {
		t.Fatalf("allowlisted host must pass, got %v", err)
	}
	if !policy.AllowsHost("DEMO") {
		t.Fatal("expected case-insensitive allowlist")
	}
}

func TestDialContextBlocksPrivate(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	policy := NewPolicy(nil)
	policy.Resolve = fakeResolver("127.0.0.1")

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := policy.DialContext(context.Background(), "tcp", net.JoinHostPort("private.test", port)); !errors.Is(err, ErrPrivateAddress) {
		t.Fatalf("expected ErrPrivateAddress, got %v", err)
	}
}

func TestDialContextAllowlistedHost(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	policy := NewPolicy([]string{"127.0.0.1"})
	address := listener.Addr().String()

	conn, err := policy.DialContext(context.Background(), "tcp", address)
	if err != nil {
		t.Fatalf("expected allowlisted dial to succeed, got %v", err)
	}
	conn.Close()
}

func TestNewClientTransport(t *testing.T) {
	policy := NewPolicy(nil)
	client := NewClient(policy, 1)
	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("unexpected transport %T", client.Transport)
	}
	if transport.Proxy != nil {
		t.Fatal("scanner client must not use a proxy")
	}
}

func TestParseURLHelper(t *testing.T) {
	parsed, err := url.Parse("http://demo/")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Hostname() != "demo" {
		t.Fatalf("unexpected host %q", parsed.Hostname())
	}
}
