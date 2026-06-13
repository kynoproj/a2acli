package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/a2aproject/a2a-go/v2/a2aclient"
)

// timestampLayout is the format used to prefix verbose log lines:
// [MM/DD/YYYY HH:MM:SS.mmm] in local time.
const timestampLayout = "[01/02/2006 15:04:05.000]"

// verboseInterceptor logs the protocol method, base URL, request payload, and
// response payload (or error) for every Client call to a writer. When the
// writer is a TTY (and colors are not disabled), key markers are colorized:
// cyan for outgoing requests, green for successful responses, red for errors.
// Every emitted line is prefixed with a HH:MM:SS.mmm timestamp.
type verboseInterceptor struct {
	a2aclient.PassthroughInterceptor
	w     io.Writer
	color bool
	now   func() time.Time
}

func newVerboseInterceptor(w io.Writer) *verboseInterceptor {
	return newVerboseInterceptorWithMode(w, colorAuto)
}

func newVerboseInterceptorWithMode(w io.Writer, mode colorMode) *verboseInterceptor {
	return &verboseInterceptor{w: w, color: colorEnabled(w, mode), now: time.Now}
}

func (v *verboseInterceptor) timestamp() string {
	return v.now().Format(timestampLayout)
}

func (v *verboseInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, any, error) {
	_, _ = fmt.Fprintf(v.w, "%s %s %s %s\n",
		v.timestamp(),
		colorize(v.color, ansiCyan, "→"),
		req.Method,
		req.BaseURL,
	)
	if req.Payload != nil {
		v.writeJSON(colorize(v.color, ansiCyan, "  request:  "), req.Payload)
	}
	return ctx, nil, nil
}

func (v *verboseInterceptor) After(ctx context.Context, resp *a2aclient.Response) error {
	ts := v.timestamp()
	if resp.Err != nil {
		_, _ = fmt.Fprintf(v.w, "%s %s %s %s %s\n",
			ts,
			colorize(v.color, ansiRed, "←"),
			resp.Method,
			colorize(v.color, ansiRed, "ERROR"),
			resp.Err,
		)
		return nil
	}
	_, _ = fmt.Fprintf(v.w, "%s %s %s %s\n",
		ts,
		colorize(v.color, ansiGreen, "←"),
		resp.Method,
		resp.BaseURL,
	)
	if resp.Payload != nil {
		v.writeJSON(colorize(v.color, ansiGreen, "  response: "), resp.Payload)
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
