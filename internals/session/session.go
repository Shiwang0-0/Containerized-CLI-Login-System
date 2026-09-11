package session

import "time"

// Session represents an authenticated user's active CLI session.
// It lives only in memory, for the lifetime of the running process.
type Session struct {
	Username    string
	LoggedInAt  time.Time
	ExpiresAt   time.Time
	LastLoginAt *time.Time // the user's PREVIOUS login time (nil if this is their first ever)
}

// New creates a fresh session that expires after `timeout` from now.
func New(username string, timeout time.Duration, lastLoginAt *time.Time) *Session {
	now := time.Now()
	return &Session{
		Username:    username,
		LoggedInAt:  now,
		ExpiresAt:   now.Add(timeout),
		LastLoginAt: lastLoginAt,
	}
}

func (s *Session) IsExpired() bool {
	if s == nil {
		return true
	}
	return time.Now().After(s.ExpiresAt)
}

// Refresh extends the session's expiry from NOW, implementing an idle/sliding timeout
// any valid activity resets the clock, so the expires at time increases
func (s *Session) Refresh(timeout time.Duration) {
	if s == nil {
		return
	}
	s.ExpiresAt = time.Now().Add(timeout)
}

// TimeRemaining returns how long until expiry (can be negative if expired).
func (s *Session) TimeRemaining() time.Duration {
	if s == nil {
		return 0
	}
	return time.Until(s.ExpiresAt)
}
