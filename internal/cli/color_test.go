package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

func TestColorize(t *testing.T) {
	t.Run("enabled wraps with ANSI codes", func(t *testing.T) {
		got := colorize(true, ansiRed, "boom")
		want := ansiRed + "boom" + ansiReset
		if got != want {
			t.Errorf("colorize(true) = %q, want %q", got, want)
		}
	})
	t.Run("disabled returns the raw string", func(t *testing.T) {
		got := colorize(false, ansiRed, "boom")
		if got != "boom" {
			t.Errorf("colorize(false) = %q, want %q", got, "boom")
		}
	})
	t.Run("empty code returns the raw string", func(t *testing.T) {
		got := colorize(true, "", "boom")
		if got != "boom" {
			t.Errorf("colorize(empty code) = %q, want %q", got, "boom")
		}
	})
}

func TestColorEnabled(t *testing.T) {
	t.Run("non-file writer is never a TTY", func(t *testing.T) {
		var buf bytes.Buffer
		if colorEnabled(&buf, colorAuto) {
			t.Error("colorEnabled(bytes.Buffer, colorAuto) = true, want false")
		}
	})
	t.Run("colorOff forces disabled even for TTY-like writers", func(t *testing.T) {
		var buf bytes.Buffer
		if colorEnabled(&buf, colorOff) {
			t.Error("colorEnabled(_, colorOff) = true, want false")
		}
	})
	t.Run("NO_COLOR env disables color", func(t *testing.T) {
		t.Setenv("NO_COLOR", "1")
		var buf bytes.Buffer
		if colorEnabled(&buf, colorAuto) {
			t.Error("colorEnabled with NO_COLOR set = true, want false")
		}
	})
}

func TestErrorLabel(t *testing.T) {
	var buf bytes.Buffer
	got := ErrorLabel(&buf)
	if got != "Error:" {
		t.Errorf("ErrorLabel(buffer) = %q, want %q", got, "Error:")
	}
}

// forceColorInterceptor returns a verbose interceptor with color forced on,
// for testing the colorized output paths without depending on a real TTY.
func forceColorInterceptor(w *bytes.Buffer) *verboseInterceptor {
	return &verboseInterceptor{w: w, color: true, now: time.Now}
}

func TestVerboseInterceptorEmitsAnsiWhenColorOn(t *testing.T) {
	var buf bytes.Buffer
	v := forceColorInterceptor(&buf)

	req := &a2aclient.Request{
		Method:  "SendMessage",
		BaseURL: "http://agent.example",
		Payload: map[string]string{"hello": "world"},
	}
	if _, _, err := v.Before(context.Background(), req); err != nil {
		t.Fatalf("Before: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, ansiCyan) {
		t.Errorf("Before output missing cyan ANSI code: %q", out)
	}
	if !strings.Contains(out, ansiReset) {
		t.Errorf("Before output missing reset ANSI code: %q", out)
	}
}
