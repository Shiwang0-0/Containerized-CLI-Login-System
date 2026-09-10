package models

import (
	"time"
)

// User is the persisted record which actually lives in the DB.
type User struct {
	Username     string
	PasswordHash string
	CreatedAt    time.Time // time of registration
	TOTPSecret   string
	TOTPEnabled  bool
}
