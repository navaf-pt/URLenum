package output

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var Sources = []string{"gau", "waybackurls", "waymore", "urlfinder", "katana", "hakrawler", "wayback-cdx"}

type Target struct {
	Output string
	Temp   string
}

func NewTarget(path string) (Target, error) {
	if strings.TrimSpace(path) == "" {
		return Target{}, fmt.Errorf("output file cannot be empty")
	}
	if strings.ToLower(filepath.Ext(path)) != ".txt" {
		path += ".txt"
	}
	if parent := filepath.Dir(path); parent != "." {
		if err := os.MkdirAll(parent, 0755); err != nil {
			return Target{}, err
		}
	}
	temp, err := os.CreateTemp("", "urlenum-*.txt")
	if err != nil {
		return Target{}, err
	}
	if err := temp.Close(); err != nil {
		os.Remove(temp.Name())
		return Target{}, err
	}
	return Target{Output: path, Temp: temp.Name()}, nil
}

func Cleanup(target Target) error { return os.Remove(target.Temp) }

func AppendWriter(target Target) (*os.File, error) {
	return os.OpenFile(target.Temp, os.O_WRONLY|os.O_APPEND, 0600)
}

func Finalize(target Target) (int, error) {
	defer os.Remove(target.Temp)
	file, err := os.Open(target.Temp)
	if err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	seen := make(map[string]struct{})
	urls := make([]string, 0)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if _, ok := seen[line]; ok {
			continue
		}
		seen[line] = struct{}{}
		urls = append(urls, line)
	}
	if err := scanner.Err(); err != nil {
		file.Close()
		return 0, err
	}
	if err := file.Close(); err != nil {
		return 0, err
	}
	sort.Strings(urls)

	output, err := os.Create(target.Output)
	if err != nil {
		return 0, err
	}
	writer := bufio.NewWriterSize(output, 64*1024)
	for _, line := range urls {
		if _, err := fmt.Fprintln(writer, line); err != nil {
			output.Close()
			return 0, err
		}
	}
	if err := writer.Flush(); err != nil {
		output.Close()
		return 0, err
	}
	if err := output.Close(); err != nil {
		return 0, err
	}
	return len(urls), nil
}
