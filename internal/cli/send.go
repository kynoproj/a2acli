package cli

import (
	"errors"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/spf13/cobra"
)

func newSendCommand(opts *globalOptions) *cobra.Command {
	var (
		accept            []string
		historyLength     int
		returnImmediately bool
	)
	cmd := &cobra.Command{
		Use:   "send [text]",
		Short: "Send a one-shot message to the agent and print the response",
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

			msg := a2a.NewMessage(a2a.MessageRoleUser, a2a.NewTextPart(text))
			req := &a2a.SendMessageRequest{Tenant: opts.tenant, Message: msg}
			if cfg := buildSendConfig(cmd, accept, historyLength, returnImmediately); cfg != nil {
				req.Config = cfg
			}
			resp, err := client.SendMessage(ctx, req)
			if err != nil {
				return err
			}
			return printJSON(cmd.OutOrStdout(), resp)
		},
	}
	f := cmd.Flags()
	f.StringSliceVar(&accept, "accept", nil, "Accepted output MIME types (repeatable or comma-separated)")
	f.IntVar(&historyLength, "history-length", 0, "Number of history messages to include in the response")
	f.BoolVar(&returnImmediately, "return-immediately", false, "Return as soon as the task is created instead of waiting for completion")
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
