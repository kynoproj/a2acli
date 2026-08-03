package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// TestSubscribeToTaskRESTEmptyBody guards the REST-transport resubscribe path
// (`task subscribe` over --protocol rest).
//
// a2a-go v2.4.0 (issue #380/#381) fixed the REST transport so a nil-payload
// request — which SubscribeToTask is — sends NO body and NO Content-Type header
// instead of a literal `null` JSON body. Before the fix, spec-compliant servers
// rejected the resubscribe request, so this exercises the one 2.4.0 change that
// a2acli actually depends on and pins it against regressions on future bumps.
func TestSubscribeToTaskRESTEmptyBody(t *testing.T) {
	const (
		taskID  = "task-123"
		wantURL = "/tasks/" + taskID + ":subscribe"
	)

	type capturedRequest struct {
		method      string
		path        string
		contentType string
		body        string
	}
	captured := make(chan capturedRequest, 1)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		captured <- capturedRequest{
			method:      r.Method,
			path:        r.URL.Path,
			contentType: r.Header.Get("Content-Type"),
			body:        string(body),
		}

		// Respond with a minimal SSE stream carrying one Task event so the
		// client's stream parser yields a value and terminates cleanly.
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "data: {\"task\":{\"id\":%q,\"contextId\":\"ctx-1\",\"status\":{\"state\":\"working\"}}}\n\n", taskID)
	}))
	defer srv.Close()

	opts := &globalOptions{
		protocol: "rest",
		endpoint: srv.URL,
		timeout:  5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, _, err := dial(ctx, opts, io.Discard)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = client.Destroy() }()

	var gotEvent bool
	req := &a2a.SubscribeToTaskRequest{ID: a2a.TaskID(taskID)}
	for event, iterErr := range client.SubscribeToTask(ctx, req) {
		if iterErr != nil {
			t.Fatalf("SubscribeToTask yielded error: %v", iterErr)
		}
		if event == nil {
			t.Fatal("SubscribeToTask yielded a nil event")
		}
		gotEvent = true
	}
	if !gotEvent {
		t.Fatal("SubscribeToTask yielded no events")
	}

	var req0 capturedRequest
	select {
	case req0 = <-captured:
	default:
		t.Fatal("server never received the subscribe request")
	}

	if req0.path != wantURL {
		t.Errorf("subscribe path = %q, want %q", req0.path, wantURL)
	}

	// The core regression assertions: an empty-payload REST request must NOT
	// carry a `null` body, and must NOT advertise a JSON Content-Type. On
	// v2.3.1 the body was "null" with Content-Type application/json.
	if body := strings.TrimSpace(req0.body); body != "" {
		t.Errorf("subscribe request body = %q, want empty (pre-2.4.0 sent %q)", body, "null")
	}
	if req0.contentType != "" {
		t.Errorf("subscribe request Content-Type = %q, want empty for a bodyless request", req0.contentType)
	}
}
