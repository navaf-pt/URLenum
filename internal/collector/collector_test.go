package collector

import (
	"testing"
	"time"

	"github.com/navaf-pt/urlenum/internal/config"
)

func TestTimeoutFor(t *testing.T) {
	runner := New(config.Config{Timeout: 180})

	if got, want := runner.timeoutFor("gau"), 30*time.Minute; got != want {
		t.Errorf("gau timeout = %s, want %s", got, want)
	}
	if got, want := runner.timeoutFor("katana"), 15*time.Minute; got != want {
		t.Errorf("Katana timeout = %s, want %s", got, want)
	}
}
