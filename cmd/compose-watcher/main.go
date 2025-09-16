package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/andrewheberle/compose-watcher/pkg/cmd"
)

func main() {
	ctx := context.Background()
	if err := cmd.Execute(ctx, os.Args[1:]); err != nil {
		slog.Error("error during execution", "error", err)
		os.Exit(1)
	}
}
