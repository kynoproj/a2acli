package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

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
