package cli

import (
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestBuildMessageJSON(t *testing.T) {
	t.Run("full-message-from-json", func(t *testing.T) {
		src := messageSources{jsonBody: `{"role":"ROLE_USER","messageId":"m-json","parts":[{"text":"hi"}]}`}
		msg, err := buildMessage(src, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.ID != "m-json" {
			t.Errorf("ID = %q, want m-json", msg.ID)
		}
		if len(msg.Parts) != 1 {
			t.Fatalf("len(Parts) = %d, want 1", len(msg.Parts))
		}
	})

	t.Run("missing-messageId-is-generated", func(t *testing.T) {
		src := messageSources{jsonBody: `{"role":"ROLE_USER","parts":[{"text":"hi"}]}`}
		msg, err := buildMessage(src, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.ID == "" {
			t.Error("ID is empty, want a generated messageId")
		}
	})

	t.Run("flags-override-json-task-context", func(t *testing.T) {
		src := messageSources{jsonBody: `{"role":"ROLE_USER","messageId":"m-1","taskId":"json-task","contextId":"json-ctx","parts":[{"text":"hi"}]}`}
		msg, err := buildMessage(src, "flag-task", "flag-ctx")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.TaskID != "flag-task" {
			t.Errorf("TaskID = %q, want flag-task", msg.TaskID)
		}
		if msg.ContextID != "flag-ctx" {
			t.Errorf("ContextID = %q, want flag-ctx", msg.ContextID)
		}
	})

	t.Run("invalid-json", func(t *testing.T) {
		src := messageSources{jsonBody: `{not valid`}
		if _, err := buildMessage(src, "", ""); err == nil {
			t.Error("expected error for invalid --json")
		}
	})
}

func TestBuildMessageParts(t *testing.T) {
	t.Run("parts-array-becomes-user-message", func(t *testing.T) {
		src := messageSources{partsJSON: `[{"text":"one"},{"text":"two"}]`}
		msg, err := buildMessage(src, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.Role != a2a.MessageRoleUser {
			t.Errorf("Role = %q, want user", msg.Role)
		}
		if len(msg.Parts) != 2 {
			t.Fatalf("len(Parts) = %d, want 2", len(msg.Parts))
		}
		if msg.ID == "" {
			t.Error("ID is empty, want a generated messageId")
		}
	})

	t.Run("invalid-parts", func(t *testing.T) {
		src := messageSources{partsJSON: `{"not":"an array"}`}
		if _, err := buildMessage(src, "", ""); err == nil {
			t.Error("expected error for invalid --parts")
		}
	})
}

// TestBuildMessagePrecedence pins the upstream source ordering:
// --json > --parts > --file > text.
func TestBuildMessagePrecedence(t *testing.T) {
	t.Run("json-wins-over-parts-and-text", func(t *testing.T) {
		src := messageSources{
			text:      "positional",
			jsonBody:  `{"role":"ROLE_USER","messageId":"m-json","parts":[{"text":"from json"}]}`,
			partsJSON: `[{"text":"from parts"}]`,
		}
		msg, err := buildMessage(src, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.ID != "m-json" {
			t.Errorf("ID = %q, want m-json (json wins)", msg.ID)
		}
	})

	t.Run("parts-wins-over-file-and-text", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-file","parts":[{"text":"from file"}]}`)
		src := messageSources{
			text:      "positional",
			file:      path,
			partsJSON: `[{"text":"from parts"}]`,
		}
		msg, err := buildMessage(src, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		// parts produces a generated ID (not m-file), confirming it beat --file.
		if msg.ID == "m-file" {
			t.Errorf("ID = %q, want a generated id (parts should win over file)", msg.ID)
		}
		if len(msg.Parts) != 1 {
			t.Fatalf("len(Parts) = %d, want 1", len(msg.Parts))
		}
	})
}
