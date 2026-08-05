package cli

import (
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

func newSendCommand(opts *globalOptions) *cobra.Command {
	var (
		accept            []string
		historyLength     int
		returnImmediately bool
		taskID            string
		contextID         string
		src               messageSources
	)
	cmd := &cobra.Command{
		Use:   "send [text]",
		Short: "Send a one-shot message to the agent and print the response",
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
			if cfg := buildSendConfig(cmd, accept, historyLength, returnImmediately); cfg != nil {
				req.Config = cfg
			}
			resp, err := client.SendMessage(ctx, req)
			if err != nil {
				return err
			}
			return opts.renderSendResult(cmd.OutOrStdout(), resp)
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&accept, "accept", nil, "Accepted output MIME types (repeatable or comma-separated)")
	f.IntVar(&historyLength, "history-length", 0, "Number of history messages to include in the response")
	f.BoolVar(&returnImmediately, "return-immediately", false, "Return as soon as the task is created instead of waiting for completion")
	f.StringVar(&taskID, "task", "", "Task ID to continue an existing task")
	f.StringVar(&contextID, "context", "", "Context ID to associate the message with an existing conversation")
	registerMessageSourceFlags(f, &src)
	return cmd
}

// buildSendConfig assembles a SendMessageConfig from --accept, --history-length,
// and --return-immediately. Returns nil when none of them are set so the
// server's defaults apply.
func buildSendConfig(cmd *cobra.Command, accept []string, historyLength int, returnImmediately bool) *a2a.SendMessageConfig {
	f := cmd.Flags()
	if !f.Changed("accept") && !f.Changed("history-length") && !f.Changed("return-immediately") {
		return nil
	}
	cfg := &a2a.SendMessageConfig{}
	if f.Changed("accept") {
		cfg.AcceptedOutputModes = accept
	}
	if f.Changed("history-length") {
		cfg.HistoryLength = &historyLength
	}
	if f.Changed("return-immediately") {
		cfg.ReturnImmediately = returnImmediately
	}
	return cfg
}

func joinArgs(args []string) string {
	return strings.Join(args, " ")
}
