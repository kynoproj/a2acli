package cli

import (
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestShortState(t *testing.T) {
	tests := []struct {
		state a2a.TaskState
		want  string
	}{
		{a2a.TaskStateWorking, "working"},
		{a2a.TaskStateCompleted, "completed"},
		{a2a.TaskStateInputRequired, "input-required"},
		{a2a.TaskState("TASK_STATE_MYSTERY"), "TASK_STATE_MYSTERY"},
	}
	for _, tt := range tests {
		if got := shortState(tt.state); got != tt.want {
			t.Errorf("shortState(%q) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestPartsText(t *testing.T) {
	tests := []struct {
		name  string
		parts a2a.ContentParts
		want  string
	}{
		{
			name:  "multiple-text-joined-by-space",
			parts: a2a.ContentParts{a2a.NewTextPart("hello"), a2a.NewTextPart("world")},
			want:  "hello world",
		},
		{
			name:  "file-url",
			parts: a2a.ContentParts{a2a.NewFileURLPart("https://example.com/f.png", "image/png")},
			want:  "[file: https://example.com/f.png]",
		},
		{
			name:  "raw-bytes",
			parts: a2a.ContentParts{a2a.NewRawPart([]byte("abcde"))},
			want:  "[binary 5 bytes]",
		},
		{
			name:  "structured-data-as-json",
			parts: a2a.ContentParts{a2a.NewDataPart(map[string]any{"k": "v"})},
			want:  `{"k":"v"}`,
		},
		{
			name:  "mixed-text-and-file",
			parts: a2a.ContentParts{a2a.NewTextPart("see"), a2a.NewFileURLPart("http://x/y", "")},
			want:  "see [file: http://x/y]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := partsText(tt.parts); got != tt.want {
				t.Errorf("partsText = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatMessage(t *testing.T) {
	agent := &a2a.Message{Role: a2a.MessageRoleAgent, Parts: a2a.ContentParts{a2a.NewTextPart("hi there")}}
	got := formatMessage(agent)
	if got != "[agent] hi there\n" {
		t.Errorf("formatMessage = %q, want %q", got, "[agent] hi there\n")
	}

	user := &a2a.Message{Role: a2a.MessageRoleUser, Parts: a2a.ContentParts{a2a.NewTextPart("q")}}
	if got := formatMessage(user); !strings.HasPrefix(got, "[user] ") {
		t.Errorf("formatMessage user prefix = %q, want [user]", got)
	}
}

func TestFormatTask(t *testing.T) {
	task := &a2a.Task{
		ID:        "task-1",
		ContextID: "ctx-9",
		Status:    a2a.TaskStatus{State: a2a.TaskStateWorking},
		History: []*a2a.Message{
			{Role: a2a.MessageRoleUser, Parts: a2a.ContentParts{a2a.NewTextPart("ping")}},
		},
	}
	got := formatTask(task)
	for _, want := range []string{"Task:", "task-1", "ctx-9", "working", "ping"} {
		if !strings.Contains(got, want) {
			t.Errorf("formatTask output missing %q; got:\n%s", want, got)
		}
	}
}

func TestFormatTaskList(t *testing.T) {
	resp := &a2a.ListTasksResponse{
		Tasks: []*a2a.Task{
			{ID: "t1", Status: a2a.TaskStatus{State: a2a.TaskStateWorking}, ContextID: "c1"},
			{ID: "t2", Status: a2a.TaskStatus{State: a2a.TaskStateCompleted}, ContextID: "c2"},
		},
		NextPageToken: "next-abc",
	}
	got := formatTaskList(resp)
	for _, want := range []string{"ID", "STATUS", "t1", "working", "t2", "completed", "next-abc"} {
		if !strings.Contains(got, want) {
			t.Errorf("formatTaskList output missing %q; got:\n%s", want, got)
		}
	}
}

func TestFormatEvent(t *testing.T) {
	tests := []struct {
		name  string
		event a2a.Event
		want  string
	}{
		{
			name:  "status-update",
			event: &a2a.TaskStatusUpdateEvent{Status: a2a.TaskStatus{State: a2a.TaskStateWorking}},
			want:  "[status] working\n",
		},
		{
			name: "artifact-update",
			event: &a2a.TaskArtifactUpdateEvent{
				Artifact: &a2a.Artifact{Parts: a2a.ContentParts{a2a.NewTextPart("chunk")}},
			},
			want: "[artifact] chunk\n",
		},
		{
			name: "artifact-append",
			event: &a2a.TaskArtifactUpdateEvent{
				Append:   true,
				Artifact: &a2a.Artifact{Parts: a2a.ContentParts{a2a.NewTextPart("more")}},
			},
			want: "[artifact+] more\n",
		},
		{
			name:  "message",
			event: &a2a.Message{Role: a2a.MessageRoleAgent, Parts: a2a.ContentParts{a2a.NewTextPart("hi")}},
			want:  "[agent] hi\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatEvent(tt.event); got != tt.want {
				t.Errorf("formatEvent = %q, want %q", got, tt.want)
			}
		})
	}

	t.Run("task-event", func(t *testing.T) {
		task := &a2a.Task{ID: "t-1", Status: a2a.TaskStatus{State: a2a.TaskStateCompleted}}
		if got := formatEvent(task); !strings.Contains(got, "t-1") || !strings.Contains(got, "completed") {
			t.Errorf("formatEvent(task) = %q, want to contain id and state", got)
		}
	})
}

func TestFormatCard(t *testing.T) {
	card := &a2a.AgentCard{
		Name:        "Test Agent",
		Description: "does things",
		Version:     "1.2.3",
	}
	got := formatCard(card)
	for _, want := range []string{"Test Agent", "does things", "1.2.3"} {
		if !strings.Contains(got, want) {
			t.Errorf("formatCard output missing %q; got:\n%s", want, got)
		}
	}
}
