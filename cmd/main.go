// Package main is an entrypoint
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"racingMetrics/internal/service"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stderr, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	jsonConfigPath := args[1]
	eventsPath := args[2]

	logger := log.New(w, "Run error", log.LstdFlags)
	runLogService := service.NewRunLog(jsonConfigPath, eventsPath, logger)
	runLogService.RunEvents(ctx)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		runLogService.PrintResultingTable()
	}
	return nil
}
