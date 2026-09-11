package cli_handler

import (
	"fmt"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
)

func (h *Handler) RegisterUser(p *prompt.Prompt) error {

	// reading the username and password and validating them
	username := readLine(p, "Username: ", auth.ValidateUsername)
	password := readSecret(p, "Password: ", auth.ValidatePassword)

	if _, err := h.userService.Register(username, password); err != nil {
		return err
	}

	fmt.Println(style.SuccessStyle.Render("Registration successful"))
	return nil
}
