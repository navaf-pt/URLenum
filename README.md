```text
 _   _ ____  _                                      
| | | |  _ \| |    ___ _ __  _   _ _ __ ___        
| | | | |_) | |   / _ \ '_ \| | | | '_ ` _ \       
| |_| |  _ <| |__|  __/ | | | |_| | | | | | |      
 \___/|_| \_\_____|\___|_| |_|\__,_|_| |_| |_|      

						 By Navaf
```

# urlenum

`urlenum` is a fast, standalone Go CLI for authorized security reconnaissance and bug bounty URL discovery. It runs independent public URL collectors one at a time in a predictable order, preserves complete URL strings including query parameters, and writes one sorted, deduplicated `.txt` result.

This project performs URL enumeration only. It does not test vulnerabilities, fuzz parameters, scan ports, brute-force directories, attack credentials, bypass authentication, or exploit targets.

## Features

- Sequential external and native HTTP collectors with bounded per-tool workers.
- Graceful cancellation on `SIGINT` and `SIGTERM`, including child process cancellation.
- Per-collector timeout, crawl depth, and collector concurrency settings. Katana receives a bounded crawl window of three times the configured timeout while retaining that value as its per-request timeout.
- One `.txt` output file containing results from all sources and domains.
- Sorting and exact-string deduplication are always applied before writing.
- Missing optional tools are skipped while the remaining collectors continue.
- No API keys, paid services, authentication, or tokens are required.
- ANSI progress output with real collector counts.

## Supported Sources

The default sources are:

- `gau`: Wayback Machine and other public providers supported by gau.
- `waybackurls`: Wayback Machine historical URLs.
- `waymore`: URL-only mode using its public sources; response downloading is disabled.
- `urlfinder`: passive subdomain discovery from configured public sources.
- `katana`: conservative standard crawling only, with configurable depth and bounded workers, page count, response size, and request rate.
- `hakrawler`: optional conservative crawling fallback.
- Wayback Machine CDX: native `net/http` request with original URLs and query strings.

As of September 2026, `katana`, `waymore`, `gau`, and `hakrawler` have active upstream repository activity. `waybackurls` is mature and stable, but its upstream release history is older. All external tools remain optional and are checked at runtime.

## Installation

Build from this repository:

```sh
go build -o urlenum .
```

Install optional collectors using their upstream instructions. For example:

```sh
go install github.com/lc/gau/v2/cmd/gau@latest
go install github.com/tomnomnom/waybackurls@latest
go install github.com/projectdiscovery/katana/cmd/katana@latest
go install github.com/projectdiscovery/urlfinder/cmd/urlfinder@latest
go install github.com/hakluke/hakrawler@latest
pipx install git+https://github.com/xnl-h4ck3r/waymore.git
```

The executable directories must be in `PATH`. `curl` and `jq` are checked for operator visibility but are not required by the Go implementation.

## Usage

```sh
./urlenum -d example.com
./urlenum -d example.com --threads 10 --timeout 30 --depth 2
./urlenum -l domains.txt -o all-results.txt
./urlenum --check
./urlenum --version
```

Available flags:

```text
-d, --domain DOMAIN    target domain
-l, --list FILE        newline-delimited domain list
-o, --output FILE      output `.txt` file (default: all-urls.txt)
--threads NUM          worker threads used by supported collectors (default: 6)
--timeout NUM          timeout per collector request in seconds (default: 180; Katana crawl budget: 3x)
--depth NUM            crawl depth for Katana and hakrawler (default: 2)
--dedupe               compatibility flag; deduplication is always enabled
--check                report installed dependencies and exit
--quiet                suppress progress output
--verbose              reserved for expanded diagnostics
-h, --help             show help
--version              show version
```

Input values are treated as domains. Keep each list entry on its own line; blank lines and lines beginning with `#` are ignored.

## Output

For `-d example.com`, output is written to `all-urls.txt` in the current directory. Use `-o custom-name.txt` to choose another file:

```sh
./urlenum -d example.com -o example-urls.txt
```

Only the final `.txt` file is kept. URLs from all successful collectors and all requested domains are combined, sorted lexicographically, and deduplicated by exact string. URLs are not normalized or filtered, and historical URLs are retained even when they are no longer reachable.

## Cross-Compilation

The project uses only the Go standard library and can be cross-compiled with `CGO_ENABLED=0`:

```sh
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o urlenum-linux-amd64 .
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o urlenum-linux-arm64 .
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o urlenum-windows-amd64.exe .
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o urlenum-darwin-amd64 .
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o urlenum-darwin-arm64 .
```

## Responsible Use

Use `urlenum` only against domains for which you have explicit authorization, such as your own assets or an applicable bug bounty scope. Respect program rules, robots and provider rate limits, target availability, and public-source terms. The operator is responsible for authorization and all use of the collected URLs.
# URLenum
