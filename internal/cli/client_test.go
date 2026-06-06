package cli

import (
	"net/http"
	"testing"
	"time"
)

func TestNewHTTPClient(t *testing.T) {
	t.Run("secure-default", func(t *testing.T) {
		c := newHTTPClient(&globalOptions{timeout: 5 * time.Second})
		if c.Timeout != 5*time.Second {
			t.Errorf("Timeout = %v, want 5s", c.Timeout)
		}
		if c.Transport != nil {
			t.Errorf("expected nil transport when --insecure is false, got %T", c.Transport)
		}
	})
	t.Run("insecure-skips-tls-verify", func(t *testing.T) {
		c := newHTTPClient(&globalOptions{timeout: 2 * time.Second, insecure: true})
		tr, ok := c.Transport.(*http.Transport)
		if !ok {
			t.Fatalf("expected *http.Transport, got %T", c.Transport)
		}
		if tr.TLSClientConfig == nil || !tr.TLSClientConfig.InsecureSkipVerify {
			t.Errorf("expected InsecureSkipVerify = true, got %+v", tr.TLSClientConfig)
		}
	})
}

func TestBuildResolveOptions(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		wantLen int
		wantErr bool
	}{
		{"nil", nil, 0, false},
		{"empty-slice", []string{}, 0, false},
		{"single", []string{"Authorization: Bearer abc"}, 1, false},
		{"trims-spaces", []string{"  X-Trace:   t-1  "}, 1, false},
		{"multiple", []string{"A: 1", "B: 2", "C: 3"}, 3, false},
		{"missing-colon", []string{"NoColonHere"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildResolveOptions(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("len(opts) = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}
