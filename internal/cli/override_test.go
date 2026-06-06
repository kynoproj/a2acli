package cli

import (
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestRewriteAuthority(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		override string
		want     string
		wantErr  bool
	}{
		{"https-host-port-replaced", "https://abc:40/path", "efg:443", "https://efg:443/path", false},
		{"https-no-port-replaced", "https://abc/path", "efg:8443", "https://efg:8443/path", false},
		{"http-preserved", "http://old.example/foo", "new.example:8080", "http://new.example:8080/foo", false},
		{"bare-hostport-replaced-wholesale", "127.0.0.1:9001", "10.0.0.1:7000", "10.0.0.1:7000", false},
		{"bare-host-replaced-wholesale", "localhost", "remote:9000", "remote:9000", false},
		{"empty-raw-returns-empty", "", "x:1", "", false},
		{"trims-input", "  https://abc:40  ", "efg:443", "https://efg:443", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewriteAuthority(tt.raw, tt.override)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHasScheme(t *testing.T) {
	tests := []struct {
		raw  string
		want bool
	}{
		{"https://example.com", true},
		{"http://x", true},
		{"grpc+tls://x", true},
		{"127.0.0.1:9001", false},
		{"localhost", false},
		{"://nope", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := hasScheme(tt.raw); got != tt.want {
				t.Errorf("hasScheme(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestApplyHostOverride(t *testing.T) {
	t.Run("empty-override-returns-same-card", func(t *testing.T) {
		card := &a2a.AgentCard{
			SupportedInterfaces: []*a2a.AgentInterface{
				{URL: "https://abc:40/api", ProtocolBinding: a2a.TransportProtocolJSONRPC},
			},
		}
		got, err := applyHostOverride(card, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != card {
			t.Errorf("expected the same pointer when override is empty")
		}
	})

	t.Run("nil-card-returns-nil", func(t *testing.T) {
		got, err := applyHostOverride(nil, "x:1")
		if err != nil || got != nil {
			t.Fatalf("got (%v, %v), want (nil, nil)", got, err)
		}
	})

	t.Run("rewrites-supported-interfaces", func(t *testing.T) {
		card := &a2a.AgentCard{
			SupportedInterfaces: []*a2a.AgentInterface{
				{URL: "https://abc:40/api", ProtocolBinding: a2a.TransportProtocolJSONRPC},
				{URL: "127.0.0.1:9001", ProtocolBinding: a2a.TransportProtocolGRPC},
			},
		}
		got, err := applyHostOverride(card, "newhost:7000")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == card {
			t.Errorf("expected a new card pointer, got the same one")
		}
		if got.SupportedInterfaces[0].URL != "https://newhost:7000/api" {
			t.Errorf("got[0].URL = %q", got.SupportedInterfaces[0].URL)
		}
		if got.SupportedInterfaces[1].URL != "newhost:7000" {
			t.Errorf("got[1].URL = %q", got.SupportedInterfaces[1].URL)
		}
		if card.SupportedInterfaces[0].URL != "https://abc:40/api" {
			t.Errorf("original card was mutated: %q", card.SupportedInterfaces[0].URL)
		}
	})
}
