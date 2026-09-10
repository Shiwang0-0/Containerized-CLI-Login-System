package cli_handler

import (
	"bufio"
	"fmt"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
)

func (h *Handler) RegisterUser(reader *bufio.Reader) error {
	fmt.Println("\n Register")

	// reading the username and password and validating them
	username := readLine(reader, "Username: ", auth.ValidateUsername)
	password := readSecret("Password: ", auth.ValidatePassword)

	if _, err := h.userService.Register(username, password); err != nil {
		return err
	}

	fmt.Println("Registration successful!")
	return nil
}
