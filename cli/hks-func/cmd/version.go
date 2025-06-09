package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Version information set by build flags
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	builtBy   = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information about hks-func CLI.`,
	Run: func(cmd *cobra.Command, args []string) {
		if output == "json" {
			fmt.Printf(`{"version":"%s","commit":"%s","date":"%s","builtBy":"%s","goVersion":"%s","platform":"%s/%s"}`,
				version, commit, date, builtBy, runtime.Version(), runtime.GOOS, runtime.GOARCH)
			fmt.Println()
		} else {
			fmt.Printf("hks-func version %s\n", version)
			if verbose {
				fmt.Printf("  Commit:   %s\n", commit)
				fmt.Printf("  Built:    %s\n", date)
				fmt.Printf("  Built by: %s\n", builtBy)
				fmt.Printf("  Go:       %s\n", runtime.Version())
				fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			}
		}
	},
}