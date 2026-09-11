package models

import (
	"errors"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrAccountLocked = ErrInvalidCredentials // same message intentionally, even if credential are same during a lockout return invalid credentials
