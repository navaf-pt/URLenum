package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	ErrCheck   = errors.New("dependency check")
	ErrVersion = errors.New("version")
)

type Config struct {
	Domains []string
	Output  string
	Threads int
	Timeout int
	Depth   int
	Dedupe  bool
	Check   bool
	Quiet   bool
	Verbose bool
}

func Parse(args []string) (Config, error) {
	var cfg Config
	var domain, list string
	fs := flag.NewFlagSet("urlenum", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	fs.StringVar(&domain, "d", "", "domain to enumerate")
	fs.StringVar(&domain, "domain", "", "domain to enumerate")
	fs.StringVar(&list, "l", "", "file containing domains")
	fs.StringVar(&list, "list", "", "file containing domains")
	fs.StringVar(&cfg.Output, "o", "all-urls.txt", "output .txt file")
	fs.StringVar(&cfg.Output, "output", "all-urls.txt", "output .txt file")
	fs.IntVar(&cfg.Threads, "threads", 6, "worker threads used by supported collectors")
	fs.IntVar(&cfg.Timeout, "timeout", 180, "timeout per collector in seconds")
	fs.IntVar(&cfg.Depth, "depth", 2, "maximum crawl depth")
	fs.BoolVar(&cfg.Dedupe, "dedupe", false, "compatibility flag; output is always sorted and deduplicated")
	fs.BoolVar(&cfg.Check, "check", false, "check external dependencies")
	fs.BoolVar(&cfg.Quiet, "quiet", false, "suppress progress output")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "show collector diagnostics")
	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "show version")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: urlenum -d DOMAIN [options]")
		fs.PrintDefaults()
	}
	if err := fs.Parse(args); err != nil {
		return cfg, err
	}
	if showVersion {
		return cfg, ErrVersion
	}
	if cfg.Check {
		return cfg, ErrCheck
	}
	if cfg.Threads < 1 || cfg.Timeout < 1 || cfg.Depth < 1 {
		return cfg, fmt.Errorf("threads, timeout, and depth must be positive")
	}
	if domain != "" {
		cfg.Domains = append(cfg.Domains, domain)
	}
	if list != "" {
		data, err := os.ReadFile(list)
		if err != nil {
			return cfg, fmt.Errorf("read domain list: %w", err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line != "" && !strings.HasPrefix(line, "#") {
				cfg.Domains = append(cfg.Domains, line)
			}
		}
	}
	if len(cfg.Domains) == 0 {
		return cfg, fmt.Errorf("provide a domain with -d or a domain list with -l")
	}
	return cfg, nil
}
