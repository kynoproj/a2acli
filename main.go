package main

import (
	"fmt"
	"os"

	"github.com/kynoproj/a2acli/internal/cli"
)

// Build-time variables populated via -ldflags. See Makefile.
var (
	version      = "dev"
	buildDate    = "unknown"
	gitCommit    = "unknown"
	gitTag       = ""
	gitTreeState = "unknown"
)

func main() {
	info := cli.VersionInfo{
		Version:      version,
		BuildDate:    buildDate,
		GitCommit:    gitCommit,
		GitTag:       gitTag,
		GitTreeState: gitTreeState,
	}
	if err := cli.NewRootCommand(info).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
