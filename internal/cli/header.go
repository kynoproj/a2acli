package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

// parsedHeader is a single "Key: Value" pair supplied via --header.
type parsedHeader struct {
	name  string
	value string
}

// parseHeaders splits each "Key: Value" string into its name and value,
// trimming surrounding whitespace. It returns an error on the first entry
// that lacks a colon separator.
func parseHeaders(headers []string) ([]parsedHeader, error) {
	if len(headers) == 0 {
		return nil, nil
	}
	out := make([]parsedHeader, 0, len(headers))
	for _, h := range headers {
		name, val, ok := strings.Cut(h, ":")
		if !ok {
			return nil, fmt.Errorf("invalid header %q: expected 'Key: Value'", h)
		}
		out = append(out, parsedHeader{
			name:  strings.TrimSpace(name),
			value: strings.TrimSpace(val),
		})
	}
	return out, nil
}

// headerInterceptor injects user-supplied headers into every outgoing protocol
// call as ServiceParams. Per the SDK binding, ServiceParams are transmitted as
// HTTP headers for jsonrpc/rest and as context metadata for gRPC, so this makes
// --header apply uniformly across all transports (including the --endpoint path).
type headerInterceptor struct {
	a2aclient.PassthroughInterceptor
	headers []parsedHeader
}

// newHeaderInterceptor builds a headerInterceptor from raw "Key: Value" flags.
// It returns nil (and no error) when no headers are supplied so callers can skip
// attaching it.
func newHeaderInterceptor(headers []string) (*headerInterceptor, error) {
	parsed, err := parseHeaders(headers)
	if err != nil {
		return nil, err
	}
	if len(parsed) == 0 {
		return nil, nil
	}
	return &headerInterceptor{headers: parsed}, nil
}

// Before appends the configured headers to the request's ServiceParams without
// mutating any shared state on the interceptor.
func (h *headerInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, any, error) {
	if req.ServiceParams == nil {
		req.ServiceParams = a2aclient.ServiceParams{}
	}
	for _, hdr := range h.headers {
		req.ServiceParams.Append(hdr.name, hdr.value)
	}
	return ctx, nil, nil
}
