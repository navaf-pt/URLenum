package ui

import (
	"fmt"
	"os/exec"
	"sync/atomic"
)

const (
	reset  = "\033[0m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
)

var quiet atomic.Bool

func SetQuiet(value bool) { quiet.Store(value) }

// Banner prints the URLenum wordmark and project attribution.
/*
	if !quiet.Load() {
		fmt.Printf(cyan + "\n" +
			"  _   _ ____  _       _                                   \n" +
			" | | | |  _ \\| |     | |__   __ _ _ __ ___  _ __         \n" +
			" | | | | |_) | |     | '_ \\ / _` | '_ ` _ \\| '_ \\        \n" +
			" | |_| |  _ <| |___  | | | | (_| | | | | | | | | | |       \n" +
			"  \\___/|_| \\_\\_____| |_| |_|\\__,_|_| |_| |_|_|_| |_|       \n" +
			reset + "\n" +
			reset + "\n")
	}
*/

// Banner prints the URLenum wordmark and project attribution.
func Banner() {
	if quiet.Load() {
		return
	}

	fmt.Printf(cyan + "\n" +
		"  _   _ ____  _                                      \n" +
		" | | | |  _ \\| |    ___ _ __  _   _ _ __ ___        \n" +
		" | | | | |_) | |   / _ \\ '_ \\| | | | '_ ` _ \\       \n" +
		" | |_| |  _ <| |__|  __/ | | | |_| | | | | | |      \n" +
		"  \\___/|_| \\_\\_____\\___|_| |_|\\__,_|_| |_| |_|      \n" +
		reset + "\n" +
		green + "                               By github.com/navaf-pt" + reset + "\n\n")
}

func Info(format string, args ...any) {
	if !quiet.Load() {
		fmt.Printf(cyan+"[+] "+format+reset+"\n", args...)
	}
}
func Success(format string, args ...any) {
	if !quiet.Load() {
		fmt.Printf(green+"[✓] "+format+reset+"\n", args...)
	}
}
func Warning(format string, args ...any) {
	if !quiet.Load() {
		fmt.Printf(yellow+"[!] "+format+reset+"\n", args...)
	}
}
func Error(format string, args ...any) {
	if !quiet.Load() {
		fmt.Printf(red+"[✗] "+format+reset+"\n", args...)
	}
}

func PrintDependencyCheck() {
	for _, name := range []string{"curl", "jq", "gau", "waybackurls", "waymore", "urlfinder", "katana", "hakrawler"} {
		if _, err := exec.LookPath(name); err == nil {
			Success("%s", name)
		} else {
			Error("%s", name)
		}
	}
}
