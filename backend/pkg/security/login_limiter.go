package security

import (
	"sync"
	"time"
)

// LoginAttempt tracks failed login attempts for a username
type LoginAttempt struct {
	Count     int
	FirstFail time.Time
	LockedAt  time.Time
}

// LoginLimiter manages login lockout state
type LoginLimiter struct {
	attempts sync.Map // key: username, value: *LoginAttempt
}

// NewLoginLimiter creates a new login limiter
func NewLoginLimiter() *LoginLimiter {
	return &LoginLimiter{}
}

// IsLocked checks if a user is currently locked out
func (l *LoginLimiter) IsLocked(username string, settings *Settings) (bool, int) {
	if !settings.LoginLockEnabled {
		return false, 0
	}

	val, ok := l.attempts.Load(username)
	if !ok {
		return false, 0
	}

	attempt := val.(*LoginAttempt)
	if attempt.LockedAt.IsZero() {
		return false, 0
	}

	lockDuration := time.Duration(settings.LoginLockDuration) * time.Minute
	remaining := time.Until(attempt.LockedAt.Add(lockDuration))
	if remaining <= 0 {
		// Lock expired, reset
		l.attempts.Delete(username)
		return false, 0
	}

	return true, int(remaining.Minutes()) + 1
}

// RecordFailure records a failed login attempt
func (l *LoginLimiter) RecordFailure(username string, settings *Settings) (locked bool, remainingMinutes int) {
	if !settings.LoginLockEnabled {
		return false, 0
	}

	val, _ := l.attempts.LoadOrStore(username, &LoginAttempt{})
	attempt := val.(*LoginAttempt)

	now := time.Now()

	// If previous attempts are older than lock duration, reset
	if !attempt.FirstFail.IsZero() {
		elapsed := now.Sub(attempt.FirstFail)
		if elapsed > time.Duration(settings.LoginLockDuration)*time.Minute {
			attempt.Count = 0
			attempt.FirstFail = time.Time{}
			attempt.LockedAt = time.Time{}
		}
	}

	if attempt.Count == 0 {
		attempt.FirstFail = now
	}
	attempt.Count++

	if attempt.Count >= settings.LoginLockAttempts {
		attempt.LockedAt = now
		return true, settings.LoginLockDuration
	}

	return false, 0
}

// RecordSuccess clears failed attempts on successful login
func (l *LoginLimiter) RecordSuccess(username string) {
	l.attempts.Delete(username)
}
