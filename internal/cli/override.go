package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
)

// applyHostOverride rewrites the host[:port] portion of every URL in the
// resolved AgentCard's SupportedInterfaces. It returns a new AgentCard with
// new AgentInterface values so the caller's card is not mutated.
//
// override is expected to be a host or host:port (no scheme).
//
// URL forms supported:
//   - scheme://host[:port]/path → authority replaced, scheme/path preserved
//   - host[:port]              → replaced wholesale (typical gRPC form)
func applyHostOverride(card *a2a.AgentCard, override string) (*a2a.AgentCard, error) {
	override = strings.TrimSpace(override)
	if override == "" || card == nil {
		return card, nil
	}

	out := *card
	if len(card.SupportedInterfaces) > 0 {
		ifaces := make([]*a2a.AgentInterface, len(card.SupportedInterfaces))
		for i, iface := range card.SupportedInterfaces {
			if iface == nil {
				continue
			}
			rewritten, err := rewriteAuthority(iface.URL, override)
			if err != nil {
				return nil, fmt.Errorf("rewrite supported interface[%d] url: %w", i, err)
			}
			copyIface := *iface
			copyIface.URL = rewritten
			ifaces[i] = &copyIface
		}
		out.SupportedInterfaces = ifaces
	}
	return &out, nil
}

// rewriteAuthority replaces the host[:port] of raw with override. If raw has no
// scheme (e.g. "127.0.0.1:9001" — the typical gRPC form), the entire value is
// replaced with override.
func rewriteAuthority(raw, override string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return raw, nil
	}
	if !hasScheme(raw) {
		return override, nil
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse %q: %w", raw, err)
	}
	u.Host = override
	return u.String(), nil
}

// hasScheme reports whether raw starts with a URL scheme followed by "://".
// This avoids treating bare "host:port" as having a scheme.
func hasScheme(raw string) bool {
	i := strings.Index(raw, "://")
	if i <= 0 {
		return false
	}
	for _, r := range raw[:i] {
		if !(r == '+' || r == '-' || r == '.' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')) {
			return false
		}
	}
	return true
}
