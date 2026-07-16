package cli

import (
	"context"
	"slices"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

func TestParseHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers []string
		want    []parsedHeader
		wantErr bool
	}{
		{"nil", nil, nil, false},
		{"empty-slice", []string{}, nil, false},
		{"single", []string{"Authorization: Bearer abc"}, []parsedHeader{{"Authorization", "Bearer abc"}}, false},
		{"trims-spaces", []string{"  X-Trace:   t-1  "}, []parsedHeader{{"X-Trace", "t-1"}}, false},
		{"empty-value", []string{"X-Empty:"}, []parsedHeader{{"X-Empty", ""}}, false},
		{"value-with-colon", []string{"X-URL: http://x:9001"}, []parsedHeader{{"X-URL", "http://x:9001"}}, false},
		{"multiple", []string{"A: 1", "B: 2"}, []parsedHeader{{"A", "1"}, {"B", "2"}}, false},
		{"missing-colon", []string{"NoColonHere"}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeaders(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d (%+v)", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestNewHeaderInterceptor(t *testing.T) {
	t.Run("nil-when-no-headers", func(t *testing.T) {
		hi, err := newHeaderInterceptor(nil)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if hi != nil {
			t.Errorf("expected nil interceptor, got %+v", hi)
		}
	})
	t.Run("error-on-invalid", func(t *testing.T) {
		if _, err := newHeaderInterceptor([]string{"bad"}); err == nil {
			t.Fatal("expected error for header without colon")
		}
	})
}

func TestHeaderInterceptorBefore(t *testing.T) {
	t.Run("injects-into-nil-service-params", func(t *testing.T) {
		hi, err := newHeaderInterceptor([]string{"Authorization: Bearer xyz", "X-Trace: t-1"})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		req := &a2aclient.Request{Method: "message/send"}
		if _, _, err := hi.Before(context.Background(), req); err != nil {
			t.Fatalf("Before returned err: %v", err)
		}
		assertHeader(t, req.ServiceParams, "Authorization", "Bearer xyz")
		assertHeader(t, req.ServiceParams, "X-Trace", "t-1")
	})

	t.Run("preserves-existing-service-params", func(t *testing.T) {
		hi, err := newHeaderInterceptor([]string{"X-Added: yes"})
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		req := &a2aclient.Request{ServiceParams: a2aclient.ServiceParams{}}
		req.ServiceParams.Append("X-Existing", "keep")
		if _, _, err := hi.Before(context.Background(), req); err != nil {
			t.Fatalf("Before returned err: %v", err)
		}
		assertHeader(t, req.ServiceParams, "X-Existing", "keep")
		assertHeader(t, req.ServiceParams, "X-Added", "yes")
	})

	t.Run("does-not-short-circuit-request", func(t *testing.T) {
		hi, _ := newHeaderInterceptor([]string{"A: 1"})
		_, result, err := hi.Before(context.Background(), &a2aclient.Request{})
		if result != nil || err != nil {
			t.Errorf("expected no short-circuit, got result=%v err=%v", result, err)
		}
	})
}

func assertHeader(t *testing.T, sp a2aclient.ServiceParams, key, want string) {
	t.Helper()
	vals := sp.Get(key)
	if len(vals) == 0 {
		t.Fatalf("header %q not present in %+v", key, sp)
	}
	if !slices.Contains(vals, want) {
		t.Errorf("header %q = %v, want to contain %q", key, vals, want)
	}
}
