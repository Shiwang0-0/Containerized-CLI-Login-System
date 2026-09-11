package session

import (
	"os"
	"strconv"
	"time"
)

const defaultSessionTimeout = 15 * time.Minute

// SessionTimeout reads SESSION_TIMEOUT_MINUTES from the environment.
// Falls back to a safe default if unset or invalid.
func SessionTimeout() time.Duration {
	raw := os.Getenv("SESSION_TIMEOUT_MINUTES")
	if raw == "" {
		return defaultSessionTimeout
	}

	minutes, err := strconv.Atoi(raw)
	if err != nil || minutes <= 0 {
		return defaultSessionTimeout
	}

	return time.Duration(minutes) * time.Minute
}
