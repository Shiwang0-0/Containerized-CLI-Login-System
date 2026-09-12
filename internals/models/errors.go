package models

import (
	"errors"

	"github.com/go-sql-driver/mysql"
)

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrAccountLocked = ErrInvalidCredentials // same message intentionally, even if credential are same during a lockout return invalid credentials
var ErrAborted = errors.New("input cancelled")

func IsDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError

	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
