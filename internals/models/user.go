package models

import (
	"database/sql"
	"time"
)

// User is the persisted record which actually lives in the DB.
type User struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time // time of registration
	TOTPSecret   string
	TOTPEnabled  bool

	FailedAttempts int // failed attempt of guessing wrong password
	LockedUntil    sql.NullTime
	LockoutLevel   int // the number of times you failed to guess, the waiting time increases exponentially then
	LastLockoutAt  sql.NullTime

	TOTPFailedAttempts int // failed attempt of guessing wrong totp
	TOTPLockedUntil    sql.NullTime
	TOTPLockoutLevel   int // the number of times you failed to guess, the waiting time increases exponentially then
	TOTPLastLockoutAt  sql.NullTime
}
