package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2aclient/agentcard"
	"github.com/spf13/cobra"
)

func newCardCommand(opts *globalOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "card",
		Short: "Fetch and print the AgentCard of an A2A server",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if strings.TrimSpace(opts.url) == "" {
				return errors.New("--url is required")
			}
			httpClient := newHTTPClient(opts)
			resolveOpts, err := buildResolveOptions(opts.header)
			if err != nil {
				return err
			}
			if opts.verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "→ AgentCard %s/.well-known/agent-card.json\n", strings.TrimRight(opts.url, "/"))
			}
			card, err := agentcard.NewResolver(httpClient).Resolve(cmd.Context(), opts.url, resolveOpts...)
			if err != nil {
				return err
			}
			if opts.verbose {
				fmt.Fprintf(cmd.ErrOrStderr(), "← AgentCard %s\n", opts.url)
			}
			card, err = applyHostOverride(card, opts.overrideHost)
			if err != nil {
				return err
			}
			return printJSON(cmd.OutOrStdout(), card)
		},
	}
}
