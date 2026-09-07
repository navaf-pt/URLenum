package collector

import (
	"testing"
	"time"

	"github.com/navaf-pt/urlenum/internal/config"
)

func TestTimeoutFor(t *testing.T) {
	runner := New(config.Config{Timeout: 180})

	if got, want := runner.timeoutFor("gau"), 180*time.Second; got != want {
		t.Errorf("non-Katana timeout = %s, want %s", got, want)
	}
	if got, want := runner.timeoutFor("katana"), 540*time.Second; got != want {
		t.Errorf("Katana timeout = %s, want %s", got, want)
	}
}
