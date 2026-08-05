package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		mode    string
		wantErr bool
	}{
		{"text", false},
		{"json", false},
		{"", true},
		{"yaml", true},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if err := validateOutput(tt.mode); (err != nil) != tt.wantErr {
				t.Errorf("validateOutput(%q) err = %v, wantErr = %v", tt.mode, err, tt.wantErr)
			}
		})
	}
}

func TestRenderTaskModes(t *testing.T) {
	task := &a2a.Task{ID: "task-1", ContextID: "ctx-1", Status: a2a.TaskStatus{State: a2a.TaskStateWorking}}

	t.Run("text", func(t *testing.T) {
		var buf bytes.Buffer
		o := &globalOptions{output: outputText}
		if err := o.renderTask(&buf, task); err != nil {
			t.Fatalf("renderTask: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, "Task:") || !strings.Contains(out, "working") {
			t.Errorf("text output missing expected content: %q", out)
		}
		if strings.Contains(out, "{") {
			t.Errorf("text output should not contain JSON braces: %q", out)
		}
	})

	t.Run("json", func(t *testing.T) {
		var buf bytes.Buffer
		o := &globalOptions{output: outputJSON}
		if err := o.renderTask(&buf, task); err != nil {
			t.Fatalf("renderTask: %v", err)
		}
		var decoded a2a.Task
		if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
			t.Fatalf("json output not valid: %v (%q)", err, buf.String())
		}
		if decoded.ID != "task-1" {
			t.Errorf("decoded ID = %q, want task-1", decoded.ID)
		}
	})
}

func TestRenderSendResultModes(t *testing.T) {
	msg := &a2a.Message{Role: a2a.MessageRoleAgent, Parts: a2a.ContentParts{a2a.NewTextPart("hi")}}

	var textBuf bytes.Buffer
	textOpts := &globalOptions{output: outputText}
	if err := textOpts.renderSendResult(&textBuf, msg); err != nil {
		t.Fatalf("renderSendResult text: %v", err)
	}
	if got := textBuf.String(); got != "[agent] hi\n" {
		t.Errorf("text render = %q, want %q", got, "[agent] hi\n")
	}

	var jsonBuf bytes.Buffer
	jsonOpts := &globalOptions{output: outputJSON}
	if err := jsonOpts.renderSendResult(&jsonBuf, msg); err != nil {
		t.Fatalf("renderSendResult json: %v", err)
	}
	if !json.Valid(jsonBuf.Bytes()) {
		t.Errorf("json render not valid JSON: %q", jsonBuf.String())
	}
}

func TestRenderCardModes(t *testing.T) {
	card := &a2a.AgentCard{Name: "Agent X", Version: "9.9.9"}

	var textBuf bytes.Buffer
	if err := (&globalOptions{output: outputText}).renderCard(&textBuf, card); err != nil {
		t.Fatalf("renderCard text: %v", err)
	}
	if got := textBuf.String(); !strings.Contains(got, "Agent X") || strings.Contains(got, "{") {
		t.Errorf("text render = %q, want plain card text", got)
	}

	var jsonBuf bytes.Buffer
	if err := (&globalOptions{output: outputJSON}).renderCard(&jsonBuf, card); err != nil {
		t.Fatalf("renderCard json: %v", err)
	}
	var decoded a2a.AgentCard
	if err := json.Unmarshal(jsonBuf.Bytes(), &decoded); err != nil {
		t.Fatalf("json render invalid: %v (%q)", err, jsonBuf.String())
	}
	if decoded.Name != "Agent X" {
		t.Errorf("decoded Name = %q, want Agent X", decoded.Name)
	}
}

func TestRenderTaskListModes(t *testing.T) {
	resp := &a2a.ListTasksResponse{
		Tasks:         []*a2a.Task{{ID: "t1", Status: a2a.TaskStatus{State: a2a.TaskStateWorking}, ContextID: "c1"}},
		NextPageToken: "tok",
	}

	var textBuf bytes.Buffer
	if err := (&globalOptions{output: outputText}).renderTaskList(&textBuf, resp); err != nil {
		t.Fatalf("renderTaskList text: %v", err)
	}
	if got := textBuf.String(); !strings.Contains(got, "t1") || !strings.Contains(got, "tok") {
		t.Errorf("text render = %q, want table with token", got)
	}

	var jsonBuf bytes.Buffer
	if err := (&globalOptions{output: outputJSON}).renderTaskList(&jsonBuf, resp); err != nil {
		t.Fatalf("renderTaskList json: %v", err)
	}
	if !json.Valid(jsonBuf.Bytes()) {
		t.Errorf("json render not valid: %q", jsonBuf.String())
	}
}

func TestRenderEventStatusUpdate(t *testing.T) {
	evt := &a2a.TaskStatusUpdateEvent{Status: a2a.TaskStatus{State: a2a.TaskStateCompleted}}

	var textBuf bytes.Buffer
	if err := (&globalOptions{output: outputText}).renderEvent(&textBuf, evt); err != nil {
		t.Fatalf("renderEvent text: %v", err)
	}
	if !strings.Contains(textBuf.String(), "completed") {
		t.Errorf("renderEvent text = %q, want to contain 'completed'", textBuf.String())
	}

	// JSON mode wraps the event in a StreamResponse envelope.
	var jsonBuf bytes.Buffer
	if err := (&globalOptions{output: outputJSON}).renderEvent(&jsonBuf, evt); err != nil {
		t.Fatalf("renderEvent json: %v", err)
	}
	if !json.Valid(jsonBuf.Bytes()) {
		t.Errorf("renderEvent json not valid: %q", jsonBuf.String())
	}
	if !strings.Contains(jsonBuf.String(), "statusUpdate") {
		t.Errorf("renderEvent json = %q, want StreamResponse envelope with statusUpdate", jsonBuf.String())
	}
}
