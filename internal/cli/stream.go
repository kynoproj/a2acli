package cli

import (
	"errors"
	"io"
	"iter"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

func newStreamCommand(opts *globalOptions) *cobra.Command {
	var (
		accept          []string
		historyLength   int
		taskID          string
		contextID       string
		pollingInterval time.Duration
		src             messageSources
	)
	cmd := &cobra.Command{
		Use:   "stream [text]",
		Short: "Send a message and stream events as they arrive",
		Args:  messageArgs(&src),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, card, err := dial(ctx, opts, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer func() { _ = client.Destroy() }()

			src.text = joinArgs(args)
			msg, err := buildMessage(src, taskID, contextID)
			if err != nil {
				return err
			}
			req := &a2a.SendMessageRequest{Tenant: opts.tenant, Message: msg}
			if cfg := buildSendConfig(cmd, accept, historyLength, false); cfg != nil {
				req.Config = cfg
			}
			out := cmd.OutOrStdout()

			err = renderEvents(opts, out, streamEvents(card, client, ctx, req))
			if err == nil || !errors.Is(err, a2a.ErrUnsupportedOperation) {
				return err
			}
			cmd.PrintErrf("streaming not listed in agent capabilities, falling back to polling (%s)\n", pollingInterval)
			return renderEvents(opts, out, handlePolling(ctx, client, req, pollingInterval))
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&accept, "accept", nil, "Accepted output MIME types (repeatable or comma-separated)")
	f.IntVar(&historyLength, "history-length", 0, "Number of history messages to include in events")
	f.StringVar(&taskID, "task", "", "Task ID to continue an existing task")
	f.StringVar(&contextID, "context", "", "Context ID to associate the message with an existing conversation")
	f.DurationVar(&pollingInterval, "polling-interval", 5*time.Second, "Duration between GetTask requests when falling back to polling")
	registerMessageSourceFlags(f, &src)
	return cmd
}

// renderEvents drains an event iterator, rendering each event via opts until
// the iterator is exhausted or yields an error.
func renderEvents(opts *globalOptions, out io.Writer, events iter.Seq2[a2a.Event, error]) error {
	for event, err := range events {
		if err != nil {
			return err
		}
		if err := opts.renderEvent(out, event); err != nil {
			return err
		}
	}
	return nil
}
