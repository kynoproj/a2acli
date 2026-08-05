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

// wireCase describes an expected on-the-wire outcome for --task/--context.
type wireCase struct {
	name          string
	args          []string
	wantTaskID    string
	wantContextID string
}

func wireCases() []wireCase {
	return []wireCase{
		{
			name:          "task-and-context",
			args:          []string{"--task", "task-42", "--context", "ctx-7"},
			wantTaskID:    "task-42",
			wantContextID: "ctx-7",
		},
		{
			name:       "task-only",
			args:       []string{"--task", "task-42"},
			wantTaskID: "task-42",
		},
		{
			name:          "context-only",
			args:          []string{"--context", "ctx-7"},
			wantContextID: "ctx-7",
		},
		{
			name: "neither-set-omits-both",
			args: nil,
		},
	}
}

// runWireCase drives the given command through an httptest server built by
// newServer, captures the outbound request body, and asserts that --task and
// --context land on the outbound Message (as taskId/contextId) and are omitted
// from the wire entirely when unset.
func runWireCase(t *testing.T, command string, tt wireCase, newServer func(bodyCh chan<- []byte) *httptest.Server) {
	t.Helper()

	bodyCh := make(chan []byte, 1)
	srv := newServer(bodyCh)
	defer srv.Close()

	root := NewRootCommand(VersionInfo{Version: "test"})
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	args := []string{
		command, "hello",
		"--endpoint", srv.URL,
		"--protocol", "jsonrpc",
	}
	args = append(args, tt.args...)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		t.Fatalf("%s failed: %v (stderr=%q)", command, err, errBuf.String())
	}

	var raw []byte
	select {
	case raw = <-bodyCh:
	default:
		t.Fatalf("server never received the %s request", command)
	}

	var env struct {
		Params struct {
			Message struct {
				TaskID    string `json:"taskId"`
				ContextID string `json:"contextId"`
			} `json:"message"`
		} `json:"params"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal request body %q: %v", raw, err)
	}
	msg := env.Params.Message

	if msg.TaskID != tt.wantTaskID {
		t.Errorf("message.taskId = %q, want %q", msg.TaskID, tt.wantTaskID)
	}
	if msg.ContextID != tt.wantContextID {
		t.Errorf("message.contextId = %q, want %q", msg.ContextID, tt.wantContextID)
	}

	// When a field is unset it must be omitted from the wire entirely, not sent
	// as an empty string, so the server starts a fresh task.
	if tt.wantTaskID == "" && strings.Contains(string(raw), `"taskId"`) {
		t.Errorf("expected taskId omitted from wire, got %q", raw)
	}
	if tt.wantContextID == "" && strings.Contains(string(raw), `"contextId"`) {
		t.Errorf("expected contextId omitted from wire, got %q", raw)
	}
}

// messageResultJSON is a JSON-RPC result wrapping a minimal agent Message.
const messageResultJSON = `{"message":{"role":"ROLE_AGENT","messageId":"m-1","content":[{"text":"ok"}]}}`

// TestSendTaskContextReachWire asserts --task/--context reach the wire for the
// non-streaming `send` command so a2acli can continue an existing task.
func TestSendTaskContextReachWire(t *testing.T) {
	newServer := func(bodyCh chan<- []byte) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			bodyCh <- body
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":` + messageResultJSON + `}`))
		}))
	}
	for _, tt := range wireCases() {
		t.Run(tt.name, func(t *testing.T) { runWireCase(t, "send", tt, newServer) })
	}
}

// TestStreamTaskContextReachWire asserts the same for the streaming `stream`
// command, whose response is delivered as a JSON-RPC-over-SSE event stream.
func TestStreamTaskContextReachWire(t *testing.T) {
	newServer := func(bodyCh chan<- []byte) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			bodyCh <- body
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			// One SSE frame carrying a JSON-RPC response with a Message event.
			_, _ = w.Write([]byte("data: " + `{"jsonrpc":"2.0","id":"1","result":` + messageResultJSON + "}\n\n"))
		}))
	}
	for _, tt := range wireCases() {
		t.Run(tt.name, func(t *testing.T) { runWireCase(t, "stream", tt, newServer) })
	}
}
