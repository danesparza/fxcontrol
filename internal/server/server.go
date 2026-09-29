// Package server manages the HTTP listener lifecycle.
package server

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

// Serve drains HTTP requests when the application context is canceled.
func Serve(ctx context.Context, listener net.Listener, handler http.Handler) error {
	server := &http.Server{Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	done := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		select {
		case <-ctx.Done():
			shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
			defer cancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Warn().Err(err).Msg("HTTP shutdown timed out; closing active connections")
				_ = server.Close()
			}
		case <-done:
		}
	}()
	log.Info().Str("address", listener.Addr().String()).Msg("HTTP API listener started")
	err := server.Serve(listener)
	close(done)
	<-stopped
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
