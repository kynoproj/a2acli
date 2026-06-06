package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := printJSON(&buf, map[string]any{"hello": "world"}); err != nil {
		t.Fatalf("printJSON: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "\"hello\": \"world\"") {
		t.Errorf("missing key/value in output: %q", out)
	}
	if !strings.HasSuffix(out, "\n") {
		t.Errorf("expected trailing newline, got %q", out)
	}
}

func TestPrintJSONUnmarshalable(t *testing.T) {
	var buf bytes.Buffer
	if err := printJSON(&buf, make(chan int)); err == nil {
		t.Errorf("expected error for unmarshalable value")
	}
}
