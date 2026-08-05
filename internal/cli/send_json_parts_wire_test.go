package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

// runSendWire drives `send` with the given extra args through a JSON-RPC
// httptest server and returns the decoded outbound message fields.
func runSendWire(t *testing.T, extraArgs ...string) struct {
	ID    string
	Parts []struct {
		Text string `json:"text"`
	}
	TaskID    string
	ContextID string
} {
	t.Helper()

	bodyCh := make(chan []byte, 1)
	srv := jsonrpcServer(bodyCh)
	defer srv.Close()

	root := NewRootCommand(VersionInfo{Version: "test"})
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	args := append([]string{
		"send",
		"--endpoint", srv.URL,
		"--protocol", "jsonrpc",
	}, extraArgs...)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		t.Fatalf("send failed: %v (stderr=%q)", err, errBuf.String())
	}

	raw := <-bodyCh
	var env struct {
		Params struct {
			Message struct {
				ID    string `json:"messageId"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
				TaskID    string `json:"taskId"`
				ContextID string `json:"contextId"`
			} `json:"message"`
		} `json:"params"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal request body %q: %v", raw, err)
	}
	m := env.Params.Message
	return struct {
		ID    string
		Parts []struct {
			Text string `json:"text"`
		}
		TaskID    string
		ContextID string
	}{m.ID, m.Parts, m.TaskID, m.ContextID}
}

func TestSendJSONReachWire(t *testing.T) {
	t.Run("json-message-reaches-wire", func(t *testing.T) {
		got := runSendWire(t, "--json", `{"role":"ROLE_USER","messageId":"m-json","parts":[{"text":"from json"}]}`)
		if got.ID != "m-json" {
			t.Errorf("messageId = %q, want m-json", got.ID)
		}
		if len(got.Parts) != 1 || got.Parts[0].Text != "from json" {
			t.Errorf("parts = %+v, want single text part %q", got.Parts, "from json")
		}
	})

	t.Run("flags-override-json-on-the-wire", func(t *testing.T) {
		got := runSendWire(t,
			"--json", `{"role":"ROLE_USER","messageId":"m-1","taskId":"json-task","contextId":"json-ctx","parts":[{"text":"hi"}]}`,
			"--task", "flag-task", "--context", "flag-ctx")
		if got.TaskID != "flag-task" {
			t.Errorf("taskId = %q, want flag-task", got.TaskID)
		}
		if got.ContextID != "flag-ctx" {
			t.Errorf("contextId = %q, want flag-ctx", got.ContextID)
		}
	})
}

func TestSendPartsReachWire(t *testing.T) {
	got := runSendWire(t, "--parts", `[{"text":"one"},{"text":"two"}]`)
	if len(got.Parts) != 2 {
		t.Fatalf("parts count = %d, want 2 (got %+v)", len(got.Parts), got.Parts)
	}
	if got.Parts[0].Text != "one" || got.Parts[1].Text != "two" {
		t.Errorf("parts = %+v, want [one two]", got.Parts)
	}
	if got.ID == "" {
		t.Error("messageId empty, want generated")
	}
}
