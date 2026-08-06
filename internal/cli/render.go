package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

const rfc3339Layout = "2006-01-02T15:04:05Z07:00"

// taskStateNames maps A2A task states to the short, human-readable names used
// in text output and accepted by the --status filter.
var taskStateNames = map[a2a.TaskState]string{
	a2a.TaskStateSubmitted:     "submitted",
	a2a.TaskStateWorking:       "working",
	a2a.TaskStateCompleted:     "completed",
	a2a.TaskStateFailed:        "failed",
	a2a.TaskStateCanceled:      "canceled",
	a2a.TaskStateRejected:      "rejected",
	a2a.TaskStateInputRequired: "input-required",
	a2a.TaskStateAuthRequired:  "auth-required",
}

// taskStatesByName is the inverse of taskStateNames, built once for O(1)
// short-name lookups in parseTaskState.
var taskStatesByName = func() map[string]a2a.TaskState {
	m := make(map[string]a2a.TaskState, len(taskStateNames))
	for state, name := range taskStateNames {
		m[name] = state
	}
	return m
}()

// shortState returns the human-readable name for a task state, falling back to
// the raw state string when unknown.
func shortState(state a2a.TaskState) string {
	if name, ok := taskStateNames[state]; ok {
		return name
	}
	return string(state)
}

// parseTaskState maps a human-readable state name (e.g. "working") to its A2A
// wire TaskState (e.g. TASK_STATE_WORKING). Input is trimmed and lowercased. It
// is the inverse of taskStateNames and returns an error for unknown names.
func parseTaskState(s string) (a2a.TaskState, error) {
	name := strings.ToLower(strings.TrimSpace(s))
	if state, ok := taskStatesByName[name]; ok {
		return state, nil
	}
	return "", fmt.Errorf("unknown task state %q: expected one of %s", s, strings.Join(taskStateNameList(), ", "))
}

// taskStateNameList returns the known short state names, sorted for stable
// error messages.
func taskStateNameList() []string {
	names := make([]string, 0, len(taskStateNames))
	for _, n := range taskStateNames {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func formatCard(card *a2a.AgentCard) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Name:         %s\n", card.Name)
	if card.Description != "" {
		fmt.Fprintf(&sb, "Description:  %s\n", card.Description)
	}
	fmt.Fprintf(&sb, "Version:      %s\n", card.Version)

	if len(card.SupportedInterfaces) > 0 {
		sb.WriteString("Interfaces:\n")
		for _, iface := range card.SupportedInterfaces {
			fmt.Fprintf(&sb, "  %-12s %s\n", iface.ProtocolBinding, iface.URL)
		}
	}

	fmt.Fprintf(&sb, "Streaming:    %v\n", card.Capabilities.Streaming)

	if len(card.Skills) > 0 {
		sb.WriteString("Skills:\n")
		for _, s := range card.Skills {
			fmt.Fprintf(&sb, "  %-20s %s\n", s.ID, s.Name)
		}
	}

	return sb.String()
}

func formatTask(task *a2a.Task) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Task:     %s\n", task.ID)
	if task.ContextID != "" {
		fmt.Fprintf(&sb, "Context:  %s\n", task.ContextID)
	}
	fmt.Fprintf(&sb, "Status:   %s", shortState(task.Status.State))
	if task.Status.Timestamp != nil {
		fmt.Fprintf(&sb, " (%s)", task.Status.Timestamp.Format(rfc3339Layout))
	}
	sb.WriteString("\n")
	if task.Status.Message != nil {
		fmt.Fprintf(&sb, "  %s\n", messageText(task.Status.Message))
	}

	if len(task.Artifacts) > 0 {
		sb.WriteString("Artifacts:\n")
		for _, art := range task.Artifacts {
			label := string(art.ID)
			if art.Name != "" {
				label = art.Name
			}
			fmt.Fprintf(&sb, "  [%s] %s\n", label, partsText(art.Parts))
		}
	}

	if len(task.History) > 0 {
		sb.WriteString("History:\n")
		for _, msg := range task.History {
			fmt.Fprintf(&sb, "  [%s] %s\n", roleName(msg.Role), messageText(msg))
		}
	}

	return sb.String()
}

func formatMessage(msg *a2a.Message) string {
	return fmt.Sprintf("[%s] %s\n", roleName(msg.Role), messageText(msg))
}

func formatTaskList(resp *a2a.ListTasksResponse) string {
	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 4, 2, ' ', 0)
	_, _ = io.WriteString(tw, "ID\tSTATUS\tCONTEXT\n")
	for _, t := range resp.Tasks {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", t.ID, shortState(t.Status.State), t.ContextID)
	}
	_ = tw.Flush()
	if resp.NextPageToken != "" {
		fmt.Fprintf(&sb, "\nNext page token: %s\n", resp.NextPageToken)
	}
	return sb.String()
}

// formatEvent renders a streaming event; it returns "" for event types that
// have no text representation.
func formatEvent(event a2a.Event) string {
	switch e := event.(type) {
	case *a2a.TaskStatusUpdateEvent:
		if e.Status.Message != nil {
			return fmt.Sprintf("[status] %s: %s\n", shortState(e.Status.State), messageText(e.Status.Message))
		}
		return fmt.Sprintf("[status] %s\n", shortState(e.Status.State))
	case *a2a.TaskArtifactUpdateEvent:
		text := partsText(e.Artifact.Parts)
		if e.Append {
			return fmt.Sprintf("[artifact+] %s\n", text)
		}
		return fmt.Sprintf("[artifact] %s\n", text)
	case *a2a.Task:
		return formatTask(e)
	case *a2a.Message:
		return formatMessage(e)
	default:
		return ""
	}
}

func roleName(role a2a.MessageRole) string {
	if role == a2a.MessageRoleAgent {
		return "agent"
	}
	return "user"
}

func messageText(msg *a2a.Message) string {
	return partsText(msg.Parts)
}

// partsText renders content parts as a single line: text verbatim, files as a
// URL marker, raw bytes as a size marker, and structured data as compact JSON.
func partsText(parts a2a.ContentParts) string {
	var sb strings.Builder
	for i, p := range parts {
		if i > 0 {
			sb.WriteString(" ")
		}
		switch {
		case p.Text() != "":
			sb.WriteString(p.Text())
		case p.URL() != "":
			fmt.Fprintf(&sb, "[file: %s]", p.URL())
		case p.Raw() != nil:
			fmt.Fprintf(&sb, "[binary %d bytes]", len(p.Raw()))
		case p.Data() != nil:
			b, err := json.Marshal(p.Data())
			if err != nil {
				fmt.Fprintf(&sb, "[data: %v]", err)
				continue
			}
			sb.Write(b)
		}
	}
	return sb.String()
}
