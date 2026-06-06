package cli

import (
	"io"
	"testing"

	"github.com/spf13/cobra"
)

// runRootForURL builds the root command, attaches a no-op probe subcommand,
// and returns the value of opts.url after PersistentPreRunE has fired.
func runRootForURL(t *testing.T, args []string) string {
	t.Helper()

	root := NewRootCommand()
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)

	var captured string
	probe := &cobra.Command{
		Use: "probe",
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Walk up to the root, then find the *globalOptions via the card
			// subcommand we know was constructed with it. Simpler: read the
			// resolved --url flag value directly.
			captured, _ = cmd.Root().PersistentFlags().GetString("url")
			return nil
		},
	}
	root.AddCommand(probe)

	root.SetArgs(append([]string{"probe"}, args...))
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	return captured
}

func TestRootURLFallback(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     string
		wantURL string
	}{
		{"flag-wins", []string{"--url", "http://flag.example"}, "http://env.example", "http://flag.example"},
		{"env-fallback", nil, "http://env.example", "http://env.example"},
		{"env-trimmed", nil, "  http://env.example  ", "http://env.example"},
		{"neither", nil, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(envServerURL, tt.env)
			if got := runRootForURL(t, tt.args); got != tt.wantURL {
				t.Errorf("url = %q, want %q", got, tt.wantURL)
			}
		})
	}
}
