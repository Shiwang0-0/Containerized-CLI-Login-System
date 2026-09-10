package cli_handler

import (
	"bufio"
	"fmt"
	"log"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
)

func (h *Handler) LoginUser(reader *bufio.Reader) error {
	fmt.Println("\nLogin")

	username := readLine(reader, "Username: ", auth.ValidateUsername)
	password := readSecret("Password: ", auth.ValidatePassword)

	user, err := h.userService.Authenticate(username, password)
	if err != nil {
		return err
	}

	log.Println(user)

	// create session on successfull login
	return nil
}
