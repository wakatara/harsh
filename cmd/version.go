package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

const version = "0.11.1"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Long:  `Print the version information for harsh along with build details.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("harsh version %s\n", version)
		fmt.Printf("go version %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	// This will be added to RootCmd in root.go
}
