package collector

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/navaf-pt/urlenum/internal/config"
	"github.com/navaf-pt/urlenum/internal/output"
	"github.com/navaf-pt/urlenum/internal/sources"
	"github.com/navaf-pt/urlenum/internal/ui"
)

type Runner struct{ cfg config.Config }

func New(cfg config.Config) Runner { return Runner{cfg: cfg} }

// timeoutFor returns the total time a collector may run. Crawlers and broad
// archive collectors may make many requests, so their total budget is larger
// than their configured per-request timeout.
func (r Runner) timeoutFor(name string) time.Duration {
	timeout := time.Duration(r.cfg.Timeout) * time.Second
	switch name {
	case "gau":
		return timeout * 10
	case "katana":
		return timeout * 5
	default:
		return timeout
	}
}

func (r Runner) Run(ctx context.Context) error {
	ui.Banner()
	target, err := output.NewTarget(r.cfg.Output)
	if err != nil {
		return err
	}
	defer func() { _ = output.Cleanup(target) }()
	for _, domain := range r.cfg.Domains {
		if err := r.runDomain(ctx, target, domain); err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
	ui.Info("Sorting and deduplicating results...")
	total, err := output.Finalize(target)
	if err != nil {
		return err
	}
	ui.Result("Complete: %d unique URLs", total)
	ui.Result("Output: %s", target.Output)
	return nil
}

func (r Runner) runDomain(ctx context.Context, target output.Target, domain string) error {
	ui.Target("Target: %s", domain)
	katanaThreads := r.cfg.Threads
	if katanaThreads > 2 {
		katanaThreads = 2
	}
	katanaBudget := int(r.timeoutFor("katana") / time.Second)
	jobs := map[string]sources.Runner{
		"gau":         sources.External("gau", "{domain}", "--threads", fmt.Sprint(r.cfg.Threads), "--timeout", fmt.Sprint(r.cfg.Timeout)),
		"waybackurls": sources.External("waybackurls"),
		"urlfinder":   sources.External("urlfinder", "-d", "{domain}", "-silent"),
		"katana":      sources.External("katana", "-u", "https://{domain}", "-d", fmt.Sprint(r.cfg.Depth), "-c", fmt.Sprint(katanaThreads), "-p", "1", "-mdp", "100", "-mrs", "1048576", "-rl", "10", "-timeout", fmt.Sprint(r.cfg.Timeout), "-ct", fmt.Sprint(katanaBudget), "-silent"),
		"hakrawler":   sources.External("hakrawler", "-d", fmt.Sprint(r.cfg.Depth), "-t", fmt.Sprint(r.cfg.Threads), "-timeout", fmt.Sprint(r.cfg.Timeout), "-subs", "-u"),
		"wayback-cdx": sources.WaybackCDX,
	}
	for _, name := range output.Sources {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		ui.Info("Running %s...", name)
		if name != "wayback-cdx" && !sources.Available(name) {
			ui.Warning("%-14s skipped (not installed)", name)
			continue
		}
		file, err := output.AppendWriter(target)
		if err != nil {
			ui.Error("%-14s failed: %v", name, err)
			continue
		}
		collectorTimeout := r.timeoutFor(name)
		jobCtx, cancel := context.WithTimeout(ctx, collectorTimeout)
		count, runErr := jobs[name](jobCtx, domain, file, sources.Config{Timeout: r.cfg.Timeout, Depth: r.cfg.Depth, Threads: r.cfg.Threads})
		cancel()
		closeErr := file.Close()
		if runErr != nil {
			if errors.Is(runErr, context.DeadlineExceeded) {
				ui.Warning("%-14s timed out after %s", name, collectorTimeout)
				continue
			}
			ui.Error("%-14s failed: %v", name, runErr)
			continue
		}
		if closeErr != nil {
			ui.Error("%-14s failed closing output: %v", name, closeErr)
			continue
		}
		ui.Success("%-14s %d URLs", name, count)
	}
	return nil
}
