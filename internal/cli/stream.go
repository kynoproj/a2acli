package cli

import (
	"errors"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

func newStreamCommand(opts *globalOptions) *cobra.Command {
	var (
		accept        []string
		historyLength int
		taskID        string
		contextID     string
	)
	cmd := &cobra.Command{
		Use:   "stream [text]",
		Short: "Send a message and stream events as they arrive",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			text := joinArgs(args)
			if text == "" {
				return errors.New("message text is empty")
			}

			ctx := cmd.Context()
			client, _, err := dial(ctx, opts, cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			defer func() { _ = client.Destroy() }()

			msg := buildUserMessage(text, taskID, contextID)
			req := &a2a.SendMessageRequest{Tenant: opts.tenant, Message: msg}
			if cfg := buildSendConfig(cmd, accept, historyLength, false); cfg != nil {
				req.Config = cfg
			}
			out := cmd.OutOrStdout()
			for event, iterErr := range client.SendStreamingMessage(ctx, req) {
				if iterErr != nil {
					return iterErr
				}
				if err := printJSON(out, event); err != nil {
					return err
				}
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&accept, "accept", nil, "Accepted output MIME types (repeatable or comma-separated)")
	f.IntVar(&historyLength, "history-length", 0, "Number of history messages to include in events")
	f.StringVar(&taskID, "task", "", "Task ID to continue an existing task")
	f.StringVar(&contextID, "context", "", "Context ID to associate the message with an existing conversation")
	return cmd
}
