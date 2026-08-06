package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// runTaskListWire drives `task list` with extraArgs through a JSON-RPC httptest
// server and returns the decoded request params. execErr captures a command
// error (for the rejection cases) instead of failing the test.
func runTaskListWire(t *testing.T, extraArgs ...string) (params struct {
	Status               string `json:"status"`
	StatusTimestampAfter string `json:"statusTimestampAfter"`
}, execErr error) {
	t.Helper()

	bodyCh := make(chan []byte, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyCh <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":{"tasks":[],"nextPageToken":""}}`))
	}))
	defer srv.Close()

	root := NewRootCommand(VersionInfo{Version: "test"})
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	args := append([]string{
		"task", "list",
		"--endpoint", srv.URL,
		"--protocol", "jsonrpc",
	}, extraArgs...)
	root.SetArgs(args)

	if execErr = root.Execute(); execErr != nil {
		return params, execErr
	}

	raw := <-bodyCh
	var env struct {
		Params struct {
			Status               string `json:"status"`
			StatusTimestampAfter string `json:"statusTimestampAfter"`
		} `json:"params"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal request body %q: %v", raw, err)
	}
	params.Status = env.Params.Status
	params.StatusTimestampAfter = env.Params.StatusTimestampAfter
	return params, nil
}

func TestTaskListStatusReachWire(t *testing.T) {
	// #27: short state name must be mapped to the wire enum, not cast raw.
	got, err := runTaskListWire(t, "--status", "working")
	if err != nil {
		t.Fatalf("task list failed: %v", err)
	}
	if got.Status != "TASK_STATE_WORKING" {
		t.Errorf("wire status = %q, want TASK_STATE_WORKING", got.Status)
	}
}

func TestTaskListStatusInvalidRejected(t *testing.T) {
	_, err := runTaskListWire(t, "--status", "bogus")
	if err == nil {
		t.Fatal("expected error for unknown --status value, got nil")
	}
	if !strings.Contains(err.Error(), "unknown task state") {
		t.Errorf("err = %v, want 'unknown task state'", err)
	}
}

func TestTaskListSinceReachWire(t *testing.T) {
	// #5: --since must be sent as statusTimestampAfter (RFC 3339).
	got, err := runTaskListWire(t, "--since", "2026-08-01T00:00:00Z")
	if err != nil {
		t.Fatalf("task list failed: %v", err)
	}
	if !strings.HasPrefix(got.StatusTimestampAfter, "2026-08-01T00:00:00") {
		t.Errorf("wire statusTimestampAfter = %q, want the 2026-08-01 timestamp", got.StatusTimestampAfter)
	}
}

func TestTaskListSinceInvalidRejected(t *testing.T) {
	_, err := runTaskListWire(t, "--since", "not-a-timestamp")
	if err == nil {
		t.Fatal("expected error for invalid --since value, got nil")
	}
	if !strings.Contains(err.Error(), "--since") {
		t.Errorf("err = %v, want '--since' parse error", err)
	}
}
