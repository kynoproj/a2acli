package cli

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	"github.com/a2aproject/a2a-go/v2/a2aclient/agentcard"
)

// newHTTPClient builds the *http.Client shared by the AgentCard resolver and
// any HTTP-backed transport (JSON-RPC, REST). When insecureSkipVerify is true,
// TLS certificate verification is disabled.
func newHTTPClient(opts *globalOptions) *http.Client {
	c := &http.Client{Timeout: opts.timeout}
	if opts.insecure {
		c.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // --insecure is opt-in
		}
	}
	return c
}

// dial resolves the AgentCard at opts.url and constructs a client using the
// transport selected by --protocol. The returned card is shared so callers can
// inspect it without an extra round trip. When opts.verbose is true, every
// protocol call is logged to verboseOut.
func dial(ctx context.Context, opts *globalOptions, verboseOut io.Writer) (*a2aclient.Client, *a2a.AgentCard, error) {
	if strings.TrimSpace(opts.url) == "" {
		return nil, nil, errors.New("--url is required")
	}

	httpClient := newHTTPClient(opts)

	transport, err := resolveTransport(opts.protocol, httpClient, opts.insecure)
	if err != nil {
		return nil, nil, err
	}

	resolveOpts, err := buildResolveOptions(opts.header)
	if err != nil {
		return nil, nil, err
	}

	if opts.verbose {
		fmt.Fprintf(verboseOut, "→ AgentCard %s/.well-known/agent-card.json\n", strings.TrimRight(opts.url, "/"))
	}
	resolver := agentcard.NewResolver(httpClient)
	card, err := resolver.Resolve(ctx, opts.url, resolveOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve agent card at %s: %w", opts.url, err)
	}
	if opts.verbose {
		fmt.Fprintf(verboseOut, "← AgentCard %s\n", opts.url)
	}
	card, err = applyHostOverride(card, opts.overrideHost)
	if err != nil {
		return nil, nil, err
	}

	factoryOpts := []a2aclient.FactoryOption{
		transport.option,
		a2aclient.WithConfig(a2aclient.Config{
			PreferredTransports: []a2a.TransportProtocol{transport.preferred},
		}),
	}
	if opts.verbose {
		factoryOpts = append(factoryOpts, a2aclient.WithCallInterceptors(newVerboseInterceptor(verboseOut)))
	}
	client, err := a2aclient.NewFromCard(ctx, card, factoryOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create a2a client: %w", err)
	}
	return client, card, nil
}

func buildResolveOptions(headers []string) ([]agentcard.ResolveOption, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	out := make([]agentcard.ResolveOption, 0, len(headers))
	for _, h := range headers {
		name, val, ok := strings.Cut(h, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q: expected 'Key: Value'", h)
		}
		out = append(out, agentcard.WithRequestHeader(strings.TrimSpace(name), strings.TrimSpace(val)))
	}
	return out, nil
}
