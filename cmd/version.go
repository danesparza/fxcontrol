package cmd

import (
	"fmt"

	"github.com/danesparza/fxcontrol/version"
	"github.com/spf13/cobra"
)

// versionCmd prints the build version and abbreviated commit ID, when available.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Shows the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\nfxcontrol version %s", version.String())

		// Append the first seven characters of the build commit ID, when set.
		if version.CommitID != "" {
			fmt.Printf(" (%s)", version.CommitID[:7])
		}

		fmt.Println(" ")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
