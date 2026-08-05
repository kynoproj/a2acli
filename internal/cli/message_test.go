package cli

import (
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestBuildUserMessage(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		taskID        string
		contextID     string
		wantTaskID    a2a.TaskID
		wantContextID string
	}{
		{
			name: "text-only",
			text: "hello",
		},
		{
			name:       "with-task",
			text:       "continue",
			taskID:     "task-1",
			wantTaskID: a2a.TaskID("task-1"),
		},
		{
			name:          "with-context",
			text:          "continue",
			contextID:     "ctx-9",
			wantContextID: "ctx-9",
		},
		{
			name:          "with-task-and-context",
			text:          "continue",
			taskID:        "task-1",
			contextID:     "ctx-9",
			wantTaskID:    a2a.TaskID("task-1"),
			wantContextID: "ctx-9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := buildUserMessage(tt.text, tt.taskID, tt.contextID)
			if msg == nil {
				t.Fatal("buildUserMessage returned nil")
			}
			if msg.Role != a2a.MessageRoleUser {
				t.Errorf("Role = %q, want %q", msg.Role, a2a.MessageRoleUser)
			}
			if len(msg.Parts) != 1 {
				t.Fatalf("len(Parts) = %d, want 1", len(msg.Parts))
			}
			if msg.TaskID != tt.wantTaskID {
				t.Errorf("TaskID = %q, want %q", msg.TaskID, tt.wantTaskID)
			}
			if msg.ContextID != tt.wantContextID {
				t.Errorf("ContextID = %q, want %q", msg.ContextID, tt.wantContextID)
			}
		})
	}
}
