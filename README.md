# A2A CLI

A command-line client for the [A2A](https://a2a-protocol.org/) (Agent-to-Agent)
Protocol, built on the official
[`a2a-go`](https://github.com/a2aproject/a2a-go) SDK.

## Install

```bash
go install github.com/kynoproj/a2acli@latest
```

Or build from source:

```bash
go build -o a2acli .
```

## Usage

All commands take `--url` (`-u`) pointing at the A2A server's base URL.
When `--url` is omitted, `a2acli` falls back to the `A2A_SERVER` environment
variable.
The AgentCard is fetched from `<url>/.well-known/agent-card.json` and the
client connects using the transport selected by `--protocol`
(default `jsonrpc`).

Supported `--protocol` values:

- `jsonrpc` — JSON-RPC over HTTP (default).
- `rest` — REST / HTTP+JSON.
- `grpc` — gRPC. Use `--insecure` for plaintext connections.

```text
a2acli [command]

Commands:
  card             Fetch and print the AgentCard
  send             Send a one-shot message and print the response
  stream           Send a message and stream events as they arrive
  task get         Fetch a task by ID
  task list        List tasks
  task cancel      Cancel a task by ID
  task subscribe   Re-subscribe to an existing task and stream events

Global flags:
  -u, --url string             Base URL of the A2A agent server (falls back to $A2A_SERVER)
  -p, --protocol string        Transport protocol: jsonrpc, rest, or grpc (default jsonrpc)
  -k, --insecure               Skip TLS verification (jsonrpc/rest) or use plaintext credentials (grpc)
      --tenant string          Optional agent-owner tenant ID applied to every request
      --timeout duration       HTTP timeout (default 30s)
  -H, --header stringArray     Extra HTTP header for the agent-card request (repeatable)

send / stream flags:
      --accept strings         Accepted output MIME types (repeatable or comma-separated)
      --history-length int     Number of history messages to include in the response
      --return-immediately     (send) Return as soon as the task is created
```

### Environment

| Variable     | Effect                                                       |
|--------------|--------------------------------------------------------------|
| `A2A_SERVER` | Default for `--url` when the flag is not provided. The flag, when set, always wins. |

```bash
export A2A_SERVER=http://127.0.0.1:9001
a2acli card                              # uses $A2A_SERVER
a2acli card --url http://other:9001      # flag wins
```

### Examples

Fetch an AgentCard:

```bash
a2acli card -u http://127.0.0.1:9001
```

Send a message:

```bash
a2acli send -u http://127.0.0.1:9001 "Hello, what can you do?"
```

Stream a message and watch task updates:

```bash
a2acli stream -u http://127.0.0.1:9001 "Summarize the latest news"
```

Inspect a task:

```bash
a2acli task get -u http://127.0.0.1:9001 <task-id>
a2acli task list -u http://127.0.0.1:9001 --status working
a2acli task cancel -u http://127.0.0.1:9001 <task-id>
a2acli task subscribe -u http://127.0.0.1:9001 <task-id>
```

Constrain the response with `SendMessageConfig` knobs:

```bash
a2acli send -u http://127.0.0.1:9001 \
  --accept application/json --history-length 5 --return-immediately \
  "Summarize this"
```

Address a tenant on multi-tenant agents:

```bash
a2acli send -u https://agent.example.com --tenant acme "Hello"
```

Pass authentication via `-H`:

```bash
a2acli card -u https://agent.example.com -H "Authorization: Bearer $TOKEN"
```

Talk to a gRPC server:

```bash
a2acli send -u http://127.0.0.1:9001 -p grpc -k "Hello"
```

Talk to a REST server:

```bash
a2acli send -u http://127.0.0.1:9001 -p rest "Hello"
```

## License

Apache-2.0. See [LICENSE](./LICENSE).
