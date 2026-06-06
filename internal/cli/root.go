// Package cli wires the a2acli subcommands.
package cli

import (
	"time"

	"github.com/spf13/cobra"
)

// globalOptions holds flags shared by all subcommands that talk to an A2A server.
type globalOptions struct {
	url      string
	timeout  time.Duration
	header   []string
	protocol string
	insecure bool
	tenant   string
}

func NewRootCommand() *cobra.Command {
	opts := &globalOptions{}

	root := &cobra.Command{
		Use:           "a2acli",
		Short:         "Command-line client for the A2A (Agent-to-Agent) protocol",
		Long:          "a2acli is a command-line client for the A2A protocol built on the official a2a-go SDK.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&opts.url, "url", "u", "", "Base URL of the A2A agent server (e.g. http://127.0.0.1:9001)")
	pf.DurationVar(&opts.timeout, "timeout", 30*time.Second, "Request timeout for the underlying HTTP client")
	pf.StringArrayVarP(&opts.header, "header", "H", nil, "Extra HTTP header to send with the agent-card request (repeatable, format: Key: Value)")
	pf.StringVarP(&opts.protocol, "protocol", "p", "jsonrpc", "Transport protocol: jsonrpc, grpc, or rest")
	pf.BoolVarP(&opts.insecure, "insecure", "k", false, "Use an insecure connection: skip TLS verification (jsonrpc/rest) or use plaintext credentials (grpc)")
	pf.StringVar(&opts.tenant, "tenant", "", "Optional agent-owner tenant ID applied to every request")

	root.AddCommand(
		newCardCommand(opts),
		newSendCommand(opts),
		newStreamCommand(opts),
		newTaskCommand(opts),
	)
	return root
}
