package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestDialDirectEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		opts      globalOptions
		wantErr   string
		wantCard  bool // expect non-nil card
		wantClose bool // expect client to be returned (and Destroy called)
	}{
		{
			name: "jsonrpc-direct",
			opts: globalOptions{
				endpoint: "http://127.0.0.1:9001",
				protocol: "jsonrpc",
				timeout:  5 * time.Second,
			},
			wantClose: true,
		},
		{
			name: "rest-direct",
			opts: globalOptions{
				endpoint: "http://127.0.0.1:9001",
				protocol: "rest",
				timeout:  5 * time.Second,
			},
			wantClose: true,
		},
		{
			name: "endpoint-with-override-host-rejected",
			opts: globalOptions{
				endpoint:     "http://127.0.0.1:9001",
				protocol:     "jsonrpc",
				overrideHost: "10.0.0.1:9001",
				timeout:      5 * time.Second,
			},
			wantErr: "--override-host is not supported with --endpoint",
		},
		{
			name: "endpoint-with-bad-protocol",
			opts: globalOptions{
				endpoint: "http://127.0.0.1:9001",
				protocol: "carrierpigeon",
				timeout:  5 * time.Second,
			},
			wantErr: "unknown protocol",
		},
		{
			name: "endpoint-empty-falls-back-to-url-required",
			opts: globalOptions{
				protocol: "jsonrpc",
				timeout:  5 * time.Second,
			},
			wantErr: "--url is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var verbose bytes.Buffer
			client, card, err := dial(context.Background(), &tt.opts, &verbose)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want substring %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantCard && card == nil {
				t.Errorf("expected non-nil card")
			}
			if !tt.wantCard && card != nil {
				t.Errorf("expected nil card in direct mode, got %+v", card)
			}
			if tt.wantClose {
				if client == nil {
					t.Fatalf("expected non-nil client")
				}
				if err := client.Destroy(); err != nil {
					t.Errorf("destroy: %v", err)
				}
			}
		})
	}
}

func TestCardCommandRejectsEndpoint(t *testing.T) {
	root := NewRootCommand(VersionInfo{Version: "test"})
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{
		"card",
		"--url", "http://127.0.0.1:9001",
		"--endpoint", "http://127.0.0.1:9001",
	})
	err := root.Execute()
	if err == nil {
		t.Fatalf("expected error from card command with --endpoint, got nil")
	}
	if !strings.Contains(err.Error(), "AgentCard") {
		t.Errorf("err = %v, want error referencing AgentCard", err)
	}
}

func TestDialDirectEndpointVerbose(t *testing.T) {
	opts := &globalOptions{
		endpoint: "http://127.0.0.1:9001",
		protocol: "jsonrpc",
		timeout:  5 * time.Second,
		verbose:  true,
	}
	var verbose bytes.Buffer
	client, _, err := dial(context.Background(), opts, &verbose)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = client.Destroy() }()

	got := verbose.String()
	if !strings.Contains(got, "DirectEndpoint") {
		t.Errorf("verbose log missing DirectEndpoint marker: %q", got)
	}
	if !strings.Contains(got, "http://127.0.0.1:9001") {
		t.Errorf("verbose log missing endpoint URL: %q", got)
	}
	if strings.Contains(got, "AgentCard") {
		t.Errorf("verbose log unexpectedly references AgentCard in direct mode: %q", got)
	}
}
