package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

// verboseInterceptor logs the protocol method, base URL, request payload, and
// response payload (or error) for every Client call to a writer.
type verboseInterceptor struct {
	a2aclient.PassthroughInterceptor
	w io.Writer
}

func newVerboseInterceptor(w io.Writer) *verboseInterceptor {
	return &verboseInterceptor{w: w}
}

func (v *verboseInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, any, error) {
	_, _ = fmt.Fprintf(v.w, "→ %s %s\n", req.Method, req.BaseURL)
	if req.Payload != nil {
		v.writeJSON("  request:  ", req.Payload)
	}
	return ctx, nil, nil
}

func (v *verboseInterceptor) After(ctx context.Context, resp *a2aclient.Response) error {
	if resp.Err != nil {
		_, _ = fmt.Fprintf(v.w, "← %s ERROR %s\n", resp.Method, resp.Err)
		return nil
	}
	_, _ = fmt.Fprintf(v.w, "← %s %s\n", resp.Method, resp.BaseURL)
	if resp.Payload != nil {
		v.writeJSON("  response: ", resp.Payload)
	}
	return nil
}

func (v *verboseInterceptor) writeJSON(prefix string, payload any) {
	b, err := json.Marshal(payload)
	if err != nil {
		_, _ = fmt.Fprintf(v.w, "%s<unmarshalable %T: %v>\n", prefix, payload, err)
		return
	}
	_, _ = fmt.Fprintf(v.w, "%s%s\n", prefix, b)
}
