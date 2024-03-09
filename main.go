package main

import (
	"context"
	"embarkCache/cache"
	logger2 "embarkCache/logger"
	"embarkCache/server"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

func run(ctx context.Context, stdout io.Writer, stderr io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	logger := logger2.New(stdout, stderr)
	c := cache.NewCache[string](ctx, time.Minute*30)
	srv := server.New(&logger, c)
	httpServer := &http.Server{
		Addr:    "localhost:3000",
		Handler: srv,
	}

	// start listening to server
	go func() {
		logger.Info().Msgf("listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("error listening and serving")
			cancel()
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)

	// wait for sigint to shutdown gracefully
	go func() {
		defer wg.Done()
		<-ctx.Done()
		// make a new context for the Shutdown (thanks Alessandro Rosetti)
		logger.Info().Msg("shutting down")
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			logger.Error().Err(err).Msg("error shutting down http server")
		}
	}()
	wg.Wait()
	return nil
}

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Stderr); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
