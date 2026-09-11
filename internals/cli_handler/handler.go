package cli_handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

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

func (h *Handler) Whoami(sess *session.Session) error {
	user, err := h.userService.GetByUsername(sess.Username)
	if err != nil {
		return err
	}

	fmt.Println("\n Account Details")
	fmt.Println("Username:        ", user.Username)
	fmt.Println("Registered on:   ", user.CreatedAt.Format(time.RFC1123))
	if user.TOTPEnabled {
		fmt.Println("MFA status:       enabled")
	} else {
		fmt.Println("MFA status:       disabled")
	}
	if sess.LastLoginAt != nil {
		fmt.Println("Last login:      ", sess.LastLoginAt.Format(time.RFC1123))
	} else {
		fmt.Println("Last login:       this is your first login")
	}
	fmt.Println("Session expires: ", sess.ExpiresAt.Format(time.RFC1123),
		fmt.Sprintf("(%s remaining)", sess.TimeRemaining().Round(time.Second)))
	fmt.Println()
	return nil
}
