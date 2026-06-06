package cli

import (
	"net/http"
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestResolveTransport(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		want     a2a.TransportProtocol
		wantErr  string
	}{
		{"default-empty", "", a2a.TransportProtocolJSONRPC, ""},
		{"jsonrpc", "jsonrpc", a2a.TransportProtocolJSONRPC, ""},
		{"jsonrpc-mixed-case", "JsonRPC", a2a.TransportProtocolJSONRPC, ""},
		{"json-rpc-dashed", "json-rpc", a2a.TransportProtocolJSONRPC, ""},
		{"grpc", "grpc", a2a.TransportProtocolGRPC, ""},
		{"grpc-upper", "GRPC", a2a.TransportProtocolGRPC, ""},
		{"rest", "rest", a2a.TransportProtocolHTTPJSON, ""},
		{"httpjson-alias", "HTTP+JSON", a2a.TransportProtocolHTTPJSON, ""},
		{"unknown", "carrierpigeon", "", "unknown protocol"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTransport(tt.protocol, &http.Client{}, false)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.preferred != tt.want {
				t.Errorf("preferred = %q, want %q", got.preferred, tt.want)
			}
			if got.option == nil {
				t.Errorf("expected non-nil FactoryOption")
			}
		})
	}
}
