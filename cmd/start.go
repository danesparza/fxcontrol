package cmd

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/danesparza/fxcontrol/api"
	_ "github.com/danesparza/fxcontrol/docs"
	"github.com/danesparza/fxcontrol/internal/discovery"
	"github.com/danesparza/fxcontrol/internal/server"
	"github.com/spf13/cobra"
)

var listenAddress string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HTTP API and background FX discovery",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		listener, err := net.Listen("tcp", listenAddress)
		if err != nil {
			return err
		}
		defer listener.Close()
		cache := &discovery.Cache{}
		stopped := make(chan struct{})
		go func() { defer close(stopped); cache.Run(ctx, discovery.NetworkScanner{}) }()
		defer func() { cancel(); <-stopped }()
		return server.Serve(ctx, listener, api.NewRouter(api.Service{StartTime: time.Now(), Discovery: cache}))
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().StringVar(&listenAddress, "listen", ":3090", "HTTP listen address")
}
