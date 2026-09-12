package cli_handler

import (
	"fmt"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
)

func (h *Handler) LoginUser(p *prompt.Prompt) (username string, lastLogin *time.Time, err error) {

	username, err = readLine(p, "Username: ", auth.ValidateUsername)
	if err != nil {
		return "", nil, err
	}

	password, err := readSecret(p, "Password: ", auth.ValidatePassword)
	if err != nil {
		return "", nil, err
	}

	user, err := h.userService.Authenticate(username, password)
	if err != nil {
		return "", nil, err
	}

	// once username and password are valid, ask for TOTP and verify it
	if user.TOTPEnabled {
		fmt.Print(style.PromptStyle.Render("Authenticator code: "))

		code, err := readLine(p, "Authenticator code: ", nil)
		if err != nil {
			return "", nil, err
		}
		if err := h.userService.VerifyTOTP(username, code); err != nil {
			fmt.Println(style.ErrorStyle.Render("Login failed:"), err)
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

	fmt.Println(style.SuccessStyle.Render("Login successful"))

	// create session on successfull login
	return username, prevLogin, nil
}
