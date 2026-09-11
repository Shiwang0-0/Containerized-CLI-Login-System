package cli_handler

import (
	"fmt"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
)

func (h *Handler) LoginUser(p *prompt.Prompt) (username string, lastLogin *time.Time, err error) {
	fmt.Println("\nLogin")

	username = readLine(p, "Username: ", auth.ValidateUsername)
	password := readSecret(p, "Password: ", auth.ValidatePassword)

	user, err := h.userService.Authenticate(username, password)
	if err != nil {
		return "", nil, err
	}

	// once username and password are valid, ask for TOTP and verify it
	if user.TOTPEnabled {
		fmt.Print("Authenticator code: ")

		code := readLine(p, "Authenticator code: ", nil)
		if err := h.userService.VerifyTOTP(username, code); err != nil {
			fmt.Println("Login failed:", err)
			return "", nil, err
		}
	} else {
		// No 2FA password, success alone completes login, reset counter here.
		if err := h.userService.ResetLoginAttempts(username); err != nil {
			fmt.Println("Error reseting login attempts") // non fatal
		}
	}

	// capture the current login as last login, before overwriting it
	var prevLogin *time.Time
	if user.LastLoginAt.Valid {
		t := user.LastLoginAt.Time
		prevLogin = &t
	}

	_ = h.userService.RecordLogin(username)

	fmt.Println("Login successful.")

	// create session on successfull login
	return username, prevLogin, nil
}
