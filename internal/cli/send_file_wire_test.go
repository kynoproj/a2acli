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

// jsonrpcServer returns a JSON-RPC httptest server that captures the request
// body on bodyCh and replies with a minimal Message result.
func jsonrpcServer(bodyCh chan<- []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyCh <- body
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"1","result":` + messageResultJSON + `}`))
	}))
}

// TestSendFileReachWire drives `send -f <file>` end-to-end and asserts the
// file's Message reaches the wire, that --task/--context override the file's
// own task/context, and that positional text combined with --file is rejected.
func TestSendFileReachWire(t *testing.T) {
	t.Run("file-message-reaches-wire", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-file","parts":[{"text":"from a file"}]}`)

		bodyCh := make(chan []byte, 1)
		srv := jsonrpcServer(bodyCh)
		defer srv.Close()

		root := NewRootCommand(VersionInfo{Version: "test"})
		var out, errBuf bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errBuf)
		root.SetArgs([]string{
			"send",
			"--endpoint", srv.URL,
			"--protocol", "jsonrpc",
			"-f", path,
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("send -f failed: %v (stderr=%q)", err, errBuf.String())
		}

		raw := <-bodyCh
		var env struct {
			Params struct {
				Message struct {
					ID    string `json:"messageId"`
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"message"`
			} `json:"params"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("unmarshal request body %q: %v", raw, err)
		}
		if env.Params.Message.ID != "m-file" {
			t.Errorf("messageId = %q, want m-file (from file)", env.Params.Message.ID)
		}
		if len(env.Params.Message.Parts) != 1 || env.Params.Message.Parts[0].Text != "from a file" {
			t.Errorf("parts = %+v, want a single text part %q", env.Params.Message.Parts, "from a file")
		}
	})

	t.Run("flags-override-file-on-the-wire", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-file","taskId":"file-task","contextId":"file-ctx","parts":[{"text":"hi"}]}`)

		bodyCh := make(chan []byte, 1)
		srv := jsonrpcServer(bodyCh)
		defer srv.Close()

		root := NewRootCommand(VersionInfo{Version: "test"})
		var out, errBuf bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errBuf)
		root.SetArgs([]string{
			"send",
			"--endpoint", srv.URL,
			"--protocol", "jsonrpc",
			"-f", path,
			"--task", "flag-task",
			"--context", "flag-ctx",
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("send -f failed: %v (stderr=%q)", err, errBuf.String())
		}

		raw := <-bodyCh
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
		if env.Params.Message.TaskID != "flag-task" {
			t.Errorf("taskId = %q, want flag-task (flag overrides file)", env.Params.Message.TaskID)
		}
		if env.Params.Message.ContextID != "flag-ctx" {
			t.Errorf("contextId = %q, want flag-ctx (flag overrides file)", env.Params.Message.ContextID)
		}
	})

	t.Run("positional-text-with-file-rejected", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","parts":[{"text":"hi"}]}`)

		root := NewRootCommand(VersionInfo{Version: "test"})
		var out, errBuf bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errBuf)
		root.SetArgs([]string{
			"send", "some text",
			"--endpoint", "http://127.0.0.1:1",
			"--protocol", "jsonrpc",
			"-f", path,
		})
		err := root.Execute()
		if err == nil {
			t.Fatal("expected error combining positional text with --file, got nil")
		}
		// Assert it's the arg-validation error, not an incidental dial failure
		// against the bogus endpoint (Args runs before RunE, so dial is never hit).
		if !strings.Contains(err.Error(), "--file") {
			t.Errorf("err = %v, want the --file validation error", err)
		}
	})

	t.Run("stream-file-reaches-wire", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-file","parts":[{"text":"stream from file"}]}`)

		bodyCh := make(chan []byte, 1)
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			bodyCh <- body
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("data: " + `{"jsonrpc":"2.0","id":"1","result":` + messageResultJSON + "}\n\n"))
		}))
		defer srv.Close()

		root := NewRootCommand(VersionInfo{Version: "test"})
		var out, errBuf bytes.Buffer
		root.SetOut(&out)
		root.SetErr(&errBuf)
		root.SetArgs([]string{
			"stream",
			"--endpoint", srv.URL,
			"--protocol", "jsonrpc",
			"-f", path,
		})
		if err := root.Execute(); err != nil {
			t.Fatalf("stream -f failed: %v (stderr=%q)", err, errBuf.String())
		}

		raw := <-bodyCh
		var env struct {
			Params struct {
				Message struct {
					ID string `json:"messageId"`
				} `json:"message"`
			} `json:"params"`
		}
		if err := json.Unmarshal(raw, &env); err != nil {
			t.Fatalf("unmarshal request body %q: %v", raw, err)
		}
		if env.Params.Message.ID != "m-file" {
			t.Errorf("messageId = %q, want m-file (from file)", env.Params.Message.ID)
		}
	})
}
