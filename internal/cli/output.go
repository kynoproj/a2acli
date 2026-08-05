package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// Output format values for --output.
const (
	outputText = "text"
	outputJSON = "json"
)

// validateOutput checks that mode is a supported --output value.
func validateOutput(mode string) error {
	switch mode {
	case outputText, outputJSON:
		return nil
	default:
		return fmt.Errorf("unknown --output %q: expected text or json", mode)
	}
}

// printJSON writes v as indented JSON to w with a trailing newline.
func printJSON(w io.Writer, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	if _, err := w.Write(b); err != nil {
		return err
	}
	_, err = w.Write([]byte{'\n'})
	return err
}

// renderCard writes an AgentCard as JSON or human-readable text per opts.output.
func (o *globalOptions) renderCard(w io.Writer, card *a2a.AgentCard) error {
	if o.output == outputJSON {
		return printJSON(w, card)
	}
	_, err := io.WriteString(w, formatCard(card))
	return err
}

// renderTask writes a Task as JSON or human-readable text per opts.output.
func (o *globalOptions) renderTask(w io.Writer, task *a2a.Task) error {
	if o.output == outputJSON {
		return printJSON(w, task)
	}
	_, err := io.WriteString(w, formatTask(task))
	return err
}

// renderTaskList writes a ListTasksResponse as JSON or a text table.
func (o *globalOptions) renderTaskList(w io.Writer, resp *a2a.ListTasksResponse) error {
	if o.output == outputJSON {
		return printJSON(w, resp)
	}
	_, err := io.WriteString(w, formatTaskList(resp))
	return err
}

// renderSendResult writes a SendMessageResult (Task or Message) per opts.output.
func (o *globalOptions) renderSendResult(w io.Writer, result a2a.SendMessageResult) error {
	if o.output == outputJSON {
		return printJSON(w, result)
	}
	switch r := result.(type) {
	case *a2a.Task:
		_, err := io.WriteString(w, formatTask(r))
		return err
	case *a2a.Message:
		_, err := io.WriteString(w, formatMessage(r))
		return err
	default:
		return nil
	}
}

// renderEvent writes a streaming event per opts.output. In JSON mode it wraps
// the event in a StreamResponse envelope to match the server wire shape.
func (o *globalOptions) renderEvent(w io.Writer, event a2a.Event) error {
	if o.output == outputJSON {
		return printJSON(w, a2a.StreamResponse{Event: event})
	}
	s := formatEvent(event)
	if s == "" {
		return nil
	}
	_, err := io.WriteString(w, s)
	return err
}
