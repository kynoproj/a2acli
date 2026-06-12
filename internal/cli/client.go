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

// dial constructs a client using the transport selected by --protocol. When
// --endpoint is set, the AgentCard fetch is bypassed and the client is built
// directly from the supplied endpoint (useful for servers whose AgentCard is
// missing or has incorrect SupportedInterfaces); in that case the returned
// card is nil. Otherwise dial resolves the AgentCard at opts.url first. When
// opts.verbose is true, every protocol call is logged to verboseOut.
func dial(ctx context.Context, opts *globalOptions, verboseOut io.Writer) (*a2aclient.Client, *a2a.AgentCard, error) {
	httpClient := newHTTPClient(opts)

	transport, err := resolveTransport(opts.protocol, httpClient, opts.insecure, opts.plaintext)
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

	if strings.TrimSpace(opts.endpoint) != "" {
		return dialDirect(ctx, opts, transport.preferred, factoryOpts, verboseOut)
	}

	if strings.TrimSpace(opts.url) == "" {
		return nil, nil, errors.New("--url is required (or use --endpoint to bypass the AgentCard)")
	}

	resolveOpts, err := buildResolveOptions(opts.header)
	if err != nil {
		return nil, nil, err
	}

	if opts.verbose {
		_, _ = fmt.Fprintf(verboseOut, "→ AgentCard %s/.well-known/agent-card.json\n", strings.TrimRight(opts.url, "/"))
	}
	resolver := agentcard.NewResolver(httpClient)
	card, err := resolver.Resolve(ctx, opts.url, resolveOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("resolve agent card at %s: %w", opts.url, err)
	}
	if opts.verbose {
		_, _ = fmt.Fprintf(verboseOut, "← AgentCard %s\n", opts.url)
	}
	card, err = applyHostOverride(card, opts.overrideHost)
	if err != nil {
		return nil, nil, err
	}

	client, err := a2aclient.NewFromCard(ctx, card, factoryOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create a2a client: %w", err)
	}
	return client, card, nil
}

// dialDirect builds a client straight from --endpoint + --protocol, skipping
// the AgentCard resolver. It synthesizes a single AgentInterface and hands it
// to a2aclient.NewFromEndpoints. The returned card is nil because no real card
// was fetched.
func dialDirect(ctx context.Context, opts *globalOptions, protocol a2a.TransportProtocol, factoryOpts []a2aclient.FactoryOption, verboseOut io.Writer) (*a2aclient.Client, *a2a.AgentCard, error) {
	if strings.TrimSpace(opts.overrideHost) != "" {
		return nil, nil, errors.New("--override-host is not supported with --endpoint; set the endpoint URL directly")
	}
	iface := a2a.NewAgentInterface(strings.TrimSpace(opts.endpoint), protocol)
	if opts.verbose {
		_, _ = fmt.Fprintf(verboseOut, "→ DirectEndpoint %s (%s)\n", iface.URL, protocol)
	}
	client, err := a2aclient.NewFromEndpoints(ctx, []*a2a.AgentInterface{iface}, factoryOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create a2a client for endpoint %s: %w", iface.URL, err)
	}
	return client, nil, nil
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
