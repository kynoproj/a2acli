package cli

import (
	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

func newStreamCommand(opts *globalOptions) *cobra.Command {
	var (
		accept        []string
		historyLength int
		taskID        string
		contextID     string
		src           messageSources
	)
	cmd := &cobra.Command{
		Use:   "stream [text]",
		Short: "Send a message and stream events as they arrive",
		Args:  messageArgs(&src),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			client, _, err := dial(ctx, opts, cmd.ErrOrStderr())
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
			for event, iterErr := range client.SendStreamingMessage(ctx, req) {
				if iterErr != nil {
					return iterErr
				}
				if err := opts.renderEvent(out, event); err != nil {
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
	registerMessageSourceFlags(f, &src)
	return cmd
}
