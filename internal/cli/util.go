package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// humanizeSince humanizes time deltas into a user friendly format
func humanizeSince(t time.Time) string {
	if t.IsZero() {
		return "never"
	}

	delta := time.Since(t)

	switch {
	// In the case of clock skew, allow for future case
	case delta < 0:
		return "in the future"
	// Now
	case delta < time.Second:
		return "now"
	// Seconds
	case delta < time.Minute:
		return fmt.Sprintf("%ds ago", int(delta.Seconds()))
	// Minutes
	case delta < time.Hour:
		return fmt.Sprintf("%dm ago", int(delta.Minutes()))
	// Hours
	case delta < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(delta.Hours()))
	// Anything else
	default:
		return fmt.Sprintf("%dd ago", int(delta.Hours()/24))
	}
}

// boolToString converts a boolean value to 'yes' or 'no'
func boolToString(b bool) string {
	var str string
	switch b {
	case true:
		str = "yes"
	case false:
		str = "no"
	}
	return str

}

// clearScreen clears the screen on multiple OSes
func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	}
}

// ParseWatchInterval parses a watch flag string into a duration.
// If val is empty, it returns 0 (disabled).
// If val is a bare number, it's interpreted as seconds.
func ParseWatchInterval(val string) (time.Duration, error) {
	if val == "" {
		return 0, nil
	}
	// try full duration format first: 500ms, 2s, 1m
	if d, err := time.ParseDuration(val); err == nil {
		return d, nil
	}
	// try interpreting as seconds if just a number
	if d, err := time.ParseDuration(val + "s"); err == nil {
		return d, nil
	}
	return 0, fmt.Errorf("invalid watch interval: %q (examples: 2s, 5, 750ms)", val)
}

// Watch repeatedly clears the screen and executes fn at a given interval.
func Watch(interval time.Duration, fn func() error) {
	for {
		clearScreen()
		if err := fn(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
		time.Sleep(interval)
	}
}
