package cli

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/a2aproject/a2a-go/v2/a2a"
	"github.com/a2aproject/a2a-go/v2/a2aclient"
	a2agrpc "github.com/a2aproject/a2a-go/v2/a2agrpc/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// transportSetup is the SDK wiring derived from a --protocol value.
type transportSetup struct {
	// option is the FactoryOption that registers the chosen transport.
	option a2aclient.FactoryOption
	// preferred is the protocol value the factory should pick from the AgentCard.
	preferred a2a.TransportProtocol
}

// resolveTransport maps the --protocol flag to an SDK FactoryOption + a
// preferred TransportProtocol value used to pick the matching AgentInterface
// from the resolved AgentCard.
//
// plaintext applies only to gRPC: when true, the connection runs without TLS.
// It is rejected for non-gRPC protocols.
func resolveTransport(protocol string, httpClient *http.Client, insecureConn, plaintext bool) (transportSetup, error) {
	normalized := strings.ToLower(strings.TrimSpace(protocol))
	if plaintext && normalized != "grpc" {
		return transportSetup{}, fmt.Errorf("--plaintext is only supported with --protocol grpc")
	}
	switch normalized {
	case "", "jsonrpc", "json-rpc", "json_rpc":
		return transportSetup{
			option:    a2aclient.WithJSONRPCTransport(httpClient),
			preferred: a2a.TransportProtocolJSONRPC,
		}, nil
	case "rest", "http+json", "httpjson":
		return transportSetup{
			option:    a2aclient.WithRESTTransport(httpClient),
			preferred: a2a.TransportProtocolHTTPJSON,
		}, nil
	case "grpc":
		var creds credentials.TransportCredentials
		if plaintext {
			creds = insecure.NewCredentials()
		} else {
			tlsCfg := &tls.Config{}
			if insecureConn {
				tlsCfg.InsecureSkipVerify = true //nolint:gosec
			}
			creds = credentials.NewTLS(tlsCfg)
		}
		dialOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(creds),
		}
		return transportSetup{
			option:    a2agrpc.WithGRPCTransport(dialOpts...),
			preferred: a2a.TransportProtocolGRPC,
		}, nil
	default:
		return transportSetup{}, fmt.Errorf("unknown protocol %q: expected one of jsonrpc, grpc, rest", protocol)
	}
}
