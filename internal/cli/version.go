package cli

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information (set at build time via ldflags)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, commit hash, and build date of nanobanana.",
	Run: func(cmd *cobra.Command, args []string) {
		f := GetFormatter()

		if f.JSONMode {
			data := map[string]string{
				"version":    Version,
				"commit":     Commit,
				"build_date": BuildDate,
				"go_version": runtime.Version(),
				"os":         runtime.GOOS,
				"arch":       runtime.GOARCH,
			}
			f.Success("version", data, nil)
			return
		}

		fmt.Printf("nanobanana %s\n", Version)
		fmt.Printf("  commit:     %s\n", Commit)
		fmt.Printf("  built:      %s\n", BuildDate)
		fmt.Printf("  go version: %s\n", runtime.Version())
		fmt.Printf("  platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}
