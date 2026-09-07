package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/navaf/urlenum/internal/collector"
	"github.com/navaf/urlenum/internal/config"
	"github.com/navaf/urlenum/internal/ui"
)

const version = "1.0.0"

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if errors.Is(err, config.ErrVersion) {
		fmt.Printf("urlenum v%s\n", version)
		return
	}
	if errors.Is(err, config.ErrCheck) {
		ui.PrintDependencyCheck()
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	ui.SetQuiet(cfg.Quiet)
	if err := collector.New(cfg).Run(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			ui.Warning("cancelled")
			return
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
