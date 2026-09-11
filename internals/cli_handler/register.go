package cli_handler

import (
	"fmt"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
)

func (h *Handler) RegisterUser(p *prompt.Prompt) error {
	fmt.Println("\n Register")

	// reading the username and password and validating them
	username := readLine(p, "Username: ", auth.ValidateUsername)
	password := readSecret(p, "Password: ", auth.ValidatePassword)

	if _, err := h.userService.Register(username, password); err != nil {
		return err
	}

	fmt.Println("Registration successful!")
	return nil
}
