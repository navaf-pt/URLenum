package sources

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"time"
)

type Config struct{ Timeout, Depth, Threads int }
type Runner func(context.Context, string, io.Writer, Config) (int, error)

func Available(name string) bool { _, err := exec.LookPath(name); return err == nil }

func External(name string, args ...string) Runner {
	return func(ctx context.Context, domain string, writer io.Writer, cfg Config) (int, error) {
		commandArgs := make([]string, len(args))
		for i, arg := range args {
			commandArgs[i] = strings.ReplaceAll(arg, "{domain}", domain)
		}
		cmd := exec.CommandContext(ctx, name, commandArgs...)
		if name == "waybackurls" {
			cmd.Stdin = strings.NewReader(domain + "\n")
		}
		if name == "hakrawler" {
			cmd.Stdin = strings.NewReader("https://" + domain + "\n")
		}
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return 0, err
		}
		stderr := new(strings.Builder)
		cmd.Stderr = stderr
		if err := cmd.Start(); err != nil {
			return 0, err
		}
		count, scanErr := copyLines(stdout, writer)
		waitErr := cmd.Wait()
		if scanErr != nil {
			return count, scanErr
		}
		if ctx.Err() != nil {
			return count, fmt.Errorf("collector stopped: %w", ctx.Err())
		}
		if waitErr != nil {
			if stderr.Len() > 0 {
				return count, fmt.Errorf("%w: %s", waitErr, strings.TrimSpace(stderr.String()))
			}
			return count, waitErr
		}
		return count, nil
	}
}

func copyLines(reader io.Reader, writer io.Writer) (int, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "" {
			continue
		}
		if _, err := fmt.Fprintln(writer, line); err != nil {
			return count, err
		}
		count++
	}
	return count, scanner.Err()
}

func WaybackCDX(ctx context.Context, domain string, writer io.Writer, cfg Config) (int, error) {
	endpoint := "https://web.archive.org/cdx/search/cdx?url=" + url.QueryEscape("*."+domain+"/*") + "&output=txt&fl=original&collapse=urlkey"
	return fetchLines(ctx, endpoint, writer, cfg.Timeout, func(line string) (string, bool) { return line, line != "" })
}

func fetchLines(ctx context.Context, endpoint string, writer io.Writer, _ int, parse func(string) (string, bool)) (int, error) {
	response, err := getWithRetry(ctx, endpoint)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return 0, fmt.Errorf("source returned %s", response.Status)
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	count := 0
	for scanner.Scan() {
		if value, ok := parse(strings.TrimSuffix(scanner.Text(), "\r")); ok {
			if _, err := fmt.Fprintln(writer, value); err != nil {
				return count, err
			}
			count++
		}
	}
	return count, scanner.Err()
}

// getWithRetry retries transient connection failures while the collector's
// deadline remains in effect. The caller controls the total timeout via ctx.
func getWithRetry(ctx context.Context, endpoint string) (*http.Response, error) {
	client := &http.Client{}
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
		if err != nil {
			return nil, err
		}
		request.Header.Set("User-Agent", "urlenum/1.0 (+authorized reconnaissance)")
		response, err := client.Do(request)
		if err == nil {
			return response, nil
		}
		lastErr = err
		if ctx.Err() != nil || attempt == 2 {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt+1) * time.Second):
		}
	}
	return nil, lastErr
}
