package cli_handler

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
)

func (h *Handler) LoginUser(reader *bufio.Reader) (string, error) {
	fmt.Println("\nLogin")

	username := readLine(reader, "Username: ", auth.ValidateUsername)
	password := readSecret("Password: ", auth.ValidatePassword)

	user, err := h.userService.Authenticate(username, password)
	if err != nil {
		return "", err
	}

	// once username and password are valid, ask for TOTP and verify it
	if user.TOTPEnabled {
		fmt.Print("Authenticator code: ")

		code, _ := reader.ReadString('\n')
		code = strings.TrimSpace(code)

		if err := h.userService.VerifyTOTP(username, code); err != nil {
			fmt.Println("Login failed:", err)
			return "", err
		}
	} else {
		// No 2FA password, success alone completes login, reset counter here.
		if err := h.userService.ResetLoginAttempts(username); err != nil {
			fmt.Println("Error reseting login attempts") // non fatal
		}
	}

	log.Println(user)

	// create session on successfull login
	return user.Username, nil
}
