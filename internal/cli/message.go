package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// registerMessageSourceFlags registers the shared message-source flags
// (--file/-f, --json, --parts) onto f, bound to src. Used by both the send and
// stream commands so the two stay in sync as sources are added.
func registerMessageSourceFlags(f *pflag.FlagSet, src *messageSources) {
	f.StringVarP(&src.file, "file", "f", "", "Read the message from a JSON file (an a2a.Message object) instead of positional text")
	f.StringVar(&src.jsonBody, "json", "", "Raw JSON a2a.Message object to send instead of positional text")
	f.StringVar(&src.partsJSON, "parts", "", "Raw JSON array of content parts to send as a user message")
}

// messageSources holds the mutually-exclusive ways to specify an outbound
// message. At most one should be non-empty; when several are set, buildMessage
// applies the upstream precedence (jsonBody > partsJSON > file > text).
type messageSources struct {
	// text is positional message text, wrapped in a single text part.
	text string
	// file is a path to a JSON a2a.Message object (--file).
	file string
	// jsonBody is a raw JSON a2a.Message object (--json).
	jsonBody string
	// partsJSON is a raw JSON array of content parts (--parts).
	partsJSON string
}

// hasNonTextSource reports whether any source other than positional text is set.
func (s messageSources) hasNonTextSource() bool {
	return s.file != "" || s.jsonBody != "" || s.partsJSON != ""
}

// messageArgs returns a cobra positional-args validator for message-sending
// commands: positional text is required unless a non-text source (--file,
// --json, or --parts) is given, in which case no positional args are allowed.
//
// src must be a pointer to the command's flag-backed messageSources: cobra
// parses flags after this validator is captured at command-construction time
// but before it is invoked, so the closure dereferences src at validation time
// to observe the parsed values. Capturing by value would always see defaults.
func messageArgs(src *messageSources) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if src.hasNonTextSource() {
			if len(args) > 0 {
				return errors.New("cannot combine positional message text with --file, --json, or --parts")
			}
			return nil
		}
		if len(args) == 0 {
			return errors.New("provide a message as text or via --file, --json, or --parts")
		}
		return nil
	}
}

// buildMessage resolves the outbound message from its source and then applies
// the --task/--context overrides. Source precedence mirrors the upstream a2a
// SDK CLI: --json > --parts > --file > positional text; --task/--context always
// override whatever the chosen source carried.
func buildMessage(src messageSources, taskID, contextID string) (*a2a.Message, error) {
	msg, err := resolveMessageSource(src)
	if err != nil {
		return nil, err
	}
	if taskID != "" {
		msg.TaskID = a2a.TaskID(taskID)
	}
	if contextID != "" {
		msg.ContextID = contextID
	}
	return msg, nil
}

// resolveMessageSource builds the base message from the highest-precedence
// source set, without applying task/context overrides. The default branch is
// unreachable through the CLI (messageArgs rejects it before RunE runs) and
// guards direct callers such as tests.
func resolveMessageSource(src messageSources) (*a2a.Message, error) {
	switch {
	case src.jsonBody != "":
		return parseMessageJSON([]byte(src.jsonBody), "--json")
	case src.partsJSON != "":
		var parts a2a.ContentParts
		if err := json.Unmarshal([]byte(src.partsJSON), &parts); err != nil {
			return nil, fmt.Errorf("parsing --parts: %w", err)
		}
		return a2a.NewMessage(a2a.MessageRoleUser, parts...), nil
	case src.file != "":
		data, err := os.ReadFile(src.file)
		if err != nil {
			return nil, fmt.Errorf("reading message file: %w", err)
		}
		return parseMessageJSON(data, fmt.Sprintf("message file %s", src.file))
	case src.text != "":
		return a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(src.text)), nil
	default:
		return nil, errors.New("provide a message as text or via --file, --json, or --parts")
	}
}

// parseMessageJSON unmarshals a full a2a.Message from data, filling a missing
// messageId with a generated ID to match the upstream SDK CLI behavior. source
// names the origin (e.g. "--json") for error context.
func parseMessageJSON(data []byte, source string) (*a2a.Message, error) {
	msg := new(a2a.Message)
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", source, err)
	}
	if msg.ID == "" {
		msg.ID = a2a.NewMessageID()
	}
	return msg, nil
}
