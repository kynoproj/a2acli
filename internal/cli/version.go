package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// VersionInfo carries build-time metadata populated by the binary's main package
// via -ldflags. See the project Makefile.
type VersionInfo struct {
	Version      string
	BuildDate    string
	GitCommit    string
	GitTag       string
	GitTreeState string
}

func newVersionCommand(info VersionInfo) *cobra.Command {
	var short bool
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the a2acli binary version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if short {
				_, err := fmt.Fprintln(out, info.Version)
				return err
			}
			_, err := fmt.Fprintf(out,
				"a2acli:\n"+
					"  Version:      %s\n"+
					"  BuildDate:    %s\n"+
					"  GitCommit:    %s\n"+
					"  GitTag:       %s\n"+
					"  GitTreeState: %s\n"+
					"  GoVersion:    %s\n"+
					"  Platform:     %s/%s\n",
				info.Version,
				info.BuildDate,
				info.GitCommit,
				orDash(info.GitTag),
				info.GitTreeState,
				runtime.Version(),
				runtime.GOOS, runtime.GOARCH,
			)
			return err
		},
	}
	cmd.Flags().BoolVar(&short, "short", false, "Print only the version string")
	return cmd
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
