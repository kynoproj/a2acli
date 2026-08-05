package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

// messageArgs returns a cobra positional-args validator for message-sending
// commands: positional text is required unless a message file (-f) is given, in
// which case no positional args are allowed.
//
// file must be a pointer to the command's --file flag variable: cobra parses
// flags after this validator is captured at command-construction time but before
// it is invoked, so the closure dereferences *file at validation time to observe
// the parsed value. Capturing file by value would always see the flag default.
func messageArgs(file *string) cobra.PositionalArgs {
	return func(_ *cobra.Command, args []string) error {
		if *file != "" {
			if len(args) > 0 {
				return errors.New("cannot combine positional message text with --file")
			}
			return nil
		}
		if len(args) == 0 {
			return errors.New("provide a message as text or via --file")
		}
		return nil
	}
}

// buildMessage resolves the outbound message from its source (a JSON Message
// file via --file, or inline text) and then applies the --task/--context
// overrides. Precedence mirrors the upstream a2a SDK CLI: a non-empty file wins
// over text, and --task/--context always override whatever the file carried.
// Exactly one source must be provided.
func buildMessage(text, file, taskID, contextID string) (*a2a.Message, error) {
	msg, err := resolveMessageSource(text, file)
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

// resolveMessageSource builds the base message from a file or inline text,
// without applying task/context overrides. The default branch is unreachable
// through the CLI (messageArgs rejects it before RunE runs) and guards direct
// callers such as tests.
func resolveMessageSource(text, file string) (*a2a.Message, error) {
	switch {
	case file != "":
		return readMessageFile(file)
	case text != "":
		return a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(text)), nil
	default:
		return nil, errors.New("provide a message as text or via --file")
	}
}

// readMessageFile reads a JSON a2a.Message from path. A missing messageId is
// filled with a generated ID, matching the upstream SDK CLI behavior.
func readMessageFile(path string) (*a2a.Message, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading message file: %w", err)
	}
	msg := new(a2a.Message)
	if err := json.Unmarshal(data, msg); err != nil {
		return nil, fmt.Errorf("parsing message file %s: %w", path, err)
	}
	if msg.ID == "" {
		msg.ID = a2a.NewMessageID()
	}
	return msg, nil
}
