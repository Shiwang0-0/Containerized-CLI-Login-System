package cli_handler

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
)

func (h *Handler) LoginUser(reader *bufio.Reader) (username string, lastLogin *time.Time, err error) {
	fmt.Println("\nLogin")

	username = readLine(reader, "Username: ", auth.ValidateUsername)
	password := readSecret("Password: ", auth.ValidatePassword)

	user, err := h.userService.Authenticate(username, password)
	if err != nil {
		return "", nil, err
	}

	// once username and password are valid, ask for TOTP and verify it
	if user.TOTPEnabled {
		fmt.Print("Authenticator code: ")

		code, _ := reader.ReadString('\n')
		code = strings.TrimSpace(code)

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

	log.Println(user)

	// create session on successfull login
	return username, prevLogin, nil
}
