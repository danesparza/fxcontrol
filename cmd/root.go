// Package cmd defines the fxcontrol command-line interface.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd displays help and dispatches the start and version subcommands.
var rootCmd = &cobra.Command{
	Use:   "fxcontrol",
	Short: "Discover FX services on the local network through an HTTP API",
	Long: `fxcontrol discovers fxaudio, fxpixel, fxdmx, and fxtrigger services
on the local network using multicast DNS and exposes the results as JSON.

Use "fxcontrol start" to launch the HTTP API and background discovery loop.
Discovery runs at startup and every 30 seconds. GET /v1/discover/ returns
the latest cached results immediately.

Use "fxcontrol version" to display build version information.`,
}

// Execute runs the selected command and exits with status 1 if it returns an error.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
