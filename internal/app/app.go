package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"forgemate/internal/config"
	"forgemate/internal/gateway"
	"forgemate/internal/state"
	"forgemate/internal/sidecar"
)

// RunGateway starts the Go control-plane skeleton and handles graceful shutdown.
func RunGateway(ctx context.Context, cfg config.Config) error {
	layout := state.ResolveLayout(cfg.StateDir, cfg.AgentID)
	if err := state.EnsureLayout(layout); err != nil {
		return fmt.Errorf("ensure state layout: %w", err)
	}

	sidecarSupervisor := sidecar.NewSupervisor()
	sidecarSupervisor.MarkHealthy()

	httpServer := gateway.NewHTTPServer(
		func() bool { return pathExists(layout.RootDir) },
		sidecarSupervisor,
	)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpServer.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("forgemate gateway listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	sigCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	select {
	case <-sigCtx.Done():
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
