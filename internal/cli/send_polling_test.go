package cli

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// pollingServer serves an AgentCard advertising no streaming support, plus a
// JSON-RPC endpoint that returns a working Task on message/send and then a
// completed Task on the first tasks/get poll. getCalls counts GetTask calls
// so callers can assert polling actually occurred.
func pollingServer(t *testing.T) (srv *httptest.Server, getCalls *atomic.Int32) {
	t.Helper()
	getCalls = &atomic.Int32{}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/agent-card.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"name":"no-stream-agent",
			"description":"test",
			"version":"1.0.0",
			"capabilities":{"streaming":false},
			"defaultInputModes":["text"],
			"defaultOutputModes":["text"],
			"skills":[],
			"supportedInterfaces":[{"url":"http://` + r.Host + `","protocolBinding":"JSONRPC","protocolVersion":"1.0"}]
		}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var env struct {
			Method string `json:"method"`
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &env)

		w.Header().Set("Content-Type", "application/json")
		switch env.Method {
		case "SendMessage":
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":{"task":{
				"id":"task-1","contextId":"ctx-1",
				"status":{"state":"TASK_STATE_WORKING"}
			}}}`))
		case "GetTask":
			n := getCalls.Add(1)
			if n < 2 {
				_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":{
					"id":"task-1","contextId":"ctx-1",
					"status":{"state":"TASK_STATE_WORKING"}
				}}`))
				return
			}
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":{
				"id":"task-1","contextId":"ctx-1",
				"status":{"state":"TASK_STATE_COMPLETED"}
			}}`))
		default:
			t.Errorf("unexpected JSON-RPC method %q", env.Method)
			w.WriteHeader(http.StatusBadRequest)
		}
	})
	return httptest.NewServer(mux), getCalls
}

// TestSendStreamFallsBackToPollingWhenUnsupported drives `send --stream`
// against a server whose AgentCard advertises capabilities.streaming=false
// and asserts a2acli polls GetTask instead of attempting an SSE stream,
// eventually rendering the completed task.
func TestSendStreamFallsBackToPollingWhenUnsupported(t *testing.T) {
	srv, getCalls := pollingServer(t)
	defer srv.Close()

	root := NewRootCommand(VersionInfo{Version: "test"})
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	root.SetArgs([]string{
		"send", "--stream", "hello",
		"--url", srv.URL,
		"--protocol", "jsonrpc",
		"--polling-interval", "1ms",
	})

	if err := root.Execute(); err != nil {
		t.Fatalf("send --stream failed: %v (stderr=%q)", err, errBuf.String())
	}

	if !bytes.Contains(errBuf.Bytes(), []byte("falling back to polling")) {
		t.Errorf("expected fallback notice on stderr, got %q", errBuf.String())
	}
	if !bytes.Contains(out.Bytes(), []byte("task-1")) {
		t.Errorf("expected rendered task in stdout, got %q", out.String())
	}
	if n := getCalls.Load(); n != 2 {
		t.Errorf("expected exactly 2 GetTask polls, got %d", n)
	}
}
