package cli_handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/service"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/session"
	"golang.org/x/term"
)

type Handler struct {
	userService *service.UserService
	session     *session.Session
}

func NewHandler(userService *service.UserService) *Handler {
	return &Handler{userService: userService}
}

// ValidateFunc validates a single piece of input.
type ValidateFunc func(string) error

// ReadLine prompts on stdin until input passes validate.
func readLine(reader *bufio.Reader, label string, validate ValidateFunc) string {
	for {
		fmt.Print(label)
		input, err := reader.ReadString('\n') // visible input while typing
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		value := strings.TrimSpace(input)
		if err := validate(value); err != nil {
			fmt.Println("Error:", err)
			continue
		}
		return value
	}
}

// ReadSecret prompts for passwords until it is validated
func readSecret(label string, validate ValidateFunc) string {
	for {
		fmt.Print(label)
		valueBytes, err := term.ReadPassword(int(os.Stdin.Fd())) // hidden input while typing
		fmt.Println()
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		value := string(valueBytes)
		if err := validate(value); err != nil {
			fmt.Println("Error:", err)
			continue
		}
		return value
	}
}
