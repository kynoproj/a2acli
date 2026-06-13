package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

func TestVerboseInterceptor(t *testing.T) {
	var buf bytes.Buffer
	v := newVerboseInterceptor(&buf)
	ctx := context.Background()

	if _, _, err := v.Before(ctx, &a2aclient.Request{
		Method:  "SendMessage",
		BaseURL: "http://agent.example",
		Payload: map[string]string{"hello": "world"},
	}); err != nil {
		t.Fatalf("Before: %v", err)
	}
	if err := v.After(ctx, &a2aclient.Response{
		Method:  "SendMessage",
		BaseURL: "http://agent.example",
		Payload: map[string]string{"answer": "42"},
	}); err != nil {
		t.Fatalf("After: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"→ SendMessage http://agent.example",
		`"hello":"world"`,
		"← SendMessage http://agent.example",
		`"answer":"42"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestVerboseInterceptorAfterError(t *testing.T) {
	var buf bytes.Buffer
	v := newVerboseInterceptor(&buf)
	if err := v.After(context.Background(), &a2aclient.Response{
		Method: "GetTask",
		Err:    errors.New("not found"),
	}); err != nil {
		t.Fatalf("After: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "ERROR not found") {
		t.Errorf("missing ERROR line in: %q", got)
	}
}

func TestVerboseInterceptorTimestamp(t *testing.T) {
	var buf bytes.Buffer
	v := newVerboseInterceptor(&buf)
	v.now = func() time.Time {
		return time.Date(2026, 6, 12, 14, 30, 45, 123_000_000, time.UTC)
	}

	if _, _, err := v.Before(context.Background(), &a2aclient.Request{
		Method:  "SendMessage",
		BaseURL: "http://agent.example",
	}); err != nil {
		t.Fatalf("Before: %v", err)
	}
	if err := v.After(context.Background(), &a2aclient.Response{
		Method:  "SendMessage",
		BaseURL: "http://agent.example",
	}); err != nil {
		t.Fatalf("After: %v", err)
	}
	if err := v.After(context.Background(), &a2aclient.Response{
		Method: "SendMessage",
		Err:    errors.New("boom"),
	}); err != nil {
		t.Fatalf("After (error): %v", err)
	}

	out := buf.String()
	for line := range strings.SplitSeq(strings.TrimRight(out, "\n"), "\n") {
		if !strings.HasPrefix(line, "[06/12/2026 14:30:45.123] ") {
			t.Errorf("line missing expected timestamp prefix: %q", line)
		}
	}
}
