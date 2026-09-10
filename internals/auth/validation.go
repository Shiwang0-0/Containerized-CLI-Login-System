package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

const (
	usernameRules = "required,min=3,max=32,username"
	passwordRules = "required,min=12,max=72"
)

var (
	validate      = validator.New()
	usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`)
)

// custom username validator
func init() {
	_ = validate.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		return usernameRegex.MatchString(fl.Field().String())
	})
}

var usernameErrMessages = map[string]string{
	"required": "username is required",
	"min":      "username must be at least 3 characters",
	"max":      "username must be at most 32 characters",
	"username": "username must start with a letter and contain only letters, numbers, and underscores",
}

var passwordErrMessages = map[string]string{
	"required": "password is required",
	"min":      "password must be at least 12 characters",
	"max":      "password must be at most 72 characters",
}

// Credentials is the username/password pair shared by every auth request.
type Credentials struct {
	Username string
	Password string
}

// String hids the password so requests are safe to log/print.
func (c Credentials) String() string {
	return fmt.Sprintf("Username: %s, Password: [HIDDEN]", c.Username)
}

func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)

	if !utf8.ValidString(username) {
		return errors.New("username contains invalid UTF-8 characters")
	}

	return mapValidationErr(validate.Var(username, usernameRules), usernameErrMessages)
}

func ValidatePassword(password string) error {
	return mapValidationErr(validate.Var(password, passwordRules), passwordErrMessages)
}

func mapValidationErr(err error, messages map[string]string) error {
	if err == nil {
		return nil
	}

	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) || len(validationErrs) == 0 {
		return errors.New("invalid input")
	}

	if msg, ok := messages[validationErrs[0].Tag()]; ok {
		return errors.New(msg)
	}
	return errors.New("invalid input")
}
