// Package cli wires the a2acli subcommands.
package cli

import (
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// envServerURL is the environment variable consulted when --url is not
// provided on the command line.
const envServerURL = "A2A_SERVER"

// globalOptions holds flags shared by all subcommands that talk to an A2A server.
type globalOptions struct {
	url          string
	endpoint     string
	timeout      time.Duration
	header       []string
	protocol     string
	insecure     bool
	plaintext    bool
	tenant       string
	verbose      bool
	overrideHost string
}

func NewRootCommand(info VersionInfo) *cobra.Command {
	opts := &globalOptions{}

	root := &cobra.Command{
		Use:           "a2acli",
		Short:         "Command-line client for the A2A (Agent-to-Agent) protocol",
		Long:          "a2acli is a command-line client for the A2A protocol built on the official a2a-go SDK.",
		Version:       info.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			if strings.TrimSpace(opts.url) == "" {
				opts.url = strings.TrimSpace(os.Getenv(envServerURL))
			}
			return nil
		},
	}

	pf := root.PersistentFlags()
	pf.StringVarP(&opts.url, "url", "u", "", "Base URL of the A2A agent server (e.g. http://127.0.0.1:9001); falls back to $A2A_SERVER")
	pf.DurationVar(&opts.timeout, "timeout", 30*time.Second, "Request timeout for the underlying HTTP client")
	pf.StringArrayVarP(&opts.header, "header", "H", nil, "Extra HTTP header to send with the agent-card request (repeatable, format: Key: Value)")
	pf.StringVarP(&opts.protocol, "protocol", "p", "jsonrpc", "Transport protocol: jsonrpc, grpc, or rest")
	pf.BoolVarP(&opts.insecure, "insecure", "k", false, "Skip TLS certificate verification (TLS is still used for encryption)")
	pf.BoolVar(&opts.plaintext, "plaintext", false, "Disable TLS entirely (gRPC only); incompatible with other protocols")
	pf.StringVar(&opts.tenant, "tenant", "", "Optional agent-owner tenant ID applied to every request")
	pf.BoolVarP(&opts.verbose, "verbose", "v", false, "Log request URL, request body, and response body to stderr")
	pf.StringVar(&opts.overrideHost, "override-host", "", "Override the host[:port] of every URL in the resolved AgentCard (e.g. 127.0.0.1:9001)")
	pf.StringVar(&opts.endpoint, "endpoint", "", "Direct endpoint URL for the chosen --protocol; when set, the AgentCard is not fetched (useful for servers with missing/incorrect SupportedInterfaces)")

	root.AddCommand(
		newCardCommand(opts),
		newSendCommand(opts),
		newStreamCommand(opts),
		newTaskCommand(opts),
		newVersionCommand(info),
	)
	return root
}
