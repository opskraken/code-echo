// cmd/version.go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const (
	Version   = "1.0.0-beta"
	BuildDate = "2025-01-27"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show CodeEcho version information",
	Long:  `Display version information for CodeEcho CLI tool, including build details and runtime information.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("CodeEcho CLI %s\n", Version)
		fmt.Printf("Build Date: %s\n", BuildDate)
		fmt.Printf("Go Version: %s\n", runtime.Version())
		fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
