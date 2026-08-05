package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "message.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	return path
}

func TestBuildMessageText(t *testing.T) {
	msg, err := buildMessage(messageSources{text: "hello world"}, "task-1", "ctx-1")
	if err != nil {
		t.Fatalf("buildMessage: %v", err)
	}
	if msg.Role != a2a.MessageRoleUser {
		t.Errorf("Role = %q, want user", msg.Role)
	}
	if len(msg.Parts) != 1 {
		t.Fatalf("len(Parts) = %d, want 1", len(msg.Parts))
	}
	if msg.TaskID != "task-1" {
		t.Errorf("TaskID = %q, want task-1", msg.TaskID)
	}
	if msg.ContextID != "ctx-1" {
		t.Errorf("ContextID = %q, want ctx-1", msg.ContextID)
	}
}

func TestBuildMessageFile(t *testing.T) {
	t.Run("full-message-from-file", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-file","parts":[{"text":"from file"}]}`)
		msg, err := buildMessage(messageSources{file: path}, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.ID != "m-file" {
			t.Errorf("ID = %q, want m-file (preserved from file)", msg.ID)
		}
		if len(msg.Parts) != 1 {
			t.Fatalf("len(Parts) = %d, want 1", len(msg.Parts))
		}
	})

	t.Run("missing-messageId-is-generated", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","parts":[{"text":"hi"}]}`)
		msg, err := buildMessage(messageSources{file: path}, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.ID == "" {
			t.Error("ID is empty, want a generated messageId")
		}
	})

	t.Run("flags-override-file-task-context", func(t *testing.T) {
		// File carries its own taskId/contextId; --task/--context must win.
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-1","taskId":"file-task","contextId":"file-ctx","parts":[{"text":"hi"}]}`)
		msg, err := buildMessage(messageSources{file: path}, "flag-task", "flag-ctx")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.TaskID != "flag-task" {
			t.Errorf("TaskID = %q, want flag-task (flag overrides file)", msg.TaskID)
		}
		if msg.ContextID != "flag-ctx" {
			t.Errorf("ContextID = %q, want flag-ctx (flag overrides file)", msg.ContextID)
		}
	})

	t.Run("file-task-context-preserved-when-flags-unset", func(t *testing.T) {
		path := writeTempFile(t, `{"role":"ROLE_USER","messageId":"m-1","taskId":"file-task","contextId":"file-ctx","parts":[{"text":"hi"}]}`)
		msg, err := buildMessage(messageSources{file: path}, "", "")
		if err != nil {
			t.Fatalf("buildMessage: %v", err)
		}
		if msg.TaskID != "file-task" {
			t.Errorf("TaskID = %q, want file-task (preserved)", msg.TaskID)
		}
		if msg.ContextID != "file-ctx" {
			t.Errorf("ContextID = %q, want file-ctx (preserved)", msg.ContextID)
		}
	})
}

func TestBuildMessageErrors(t *testing.T) {
	t.Run("no-source", func(t *testing.T) {
		if _, err := buildMessage(messageSources{}, "", ""); err == nil {
			t.Error("expected error when neither text nor file is provided")
		}
	})

	t.Run("unreadable-file", func(t *testing.T) {
		_, err := buildMessage(messageSources{file: "/nonexistent/does-not-exist.json"}, "", "")
		if err == nil {
			t.Error("expected error for unreadable file")
		}
	})

	t.Run("invalid-json", func(t *testing.T) {
		path := writeTempFile(t, `{not valid json`)
		if _, err := buildMessage(messageSources{file: path}, "", ""); err == nil {
			t.Error("expected error for invalid JSON in file")
		}
	})
}
