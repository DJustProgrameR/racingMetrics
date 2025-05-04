package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	jsonConfigPath := args[1]
	eventsPath := args[2]

	return nil
}
