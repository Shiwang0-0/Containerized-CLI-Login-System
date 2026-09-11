package cli_handler

import (
	"fmt"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/service"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/session"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
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
func readLine(p *prompt.Prompt, label string, validate func(string) error) string {
	for {
		val, err := p.ReadLine(label)
		if err != nil {
			// treat EOF/interrupt as empty input for now; caller loops will re-prompt
			continue
		}
		if validate != nil {
			if verr := validate(val); verr != nil {
				fmt.Println("Invalid input:", verr)
				continue
			}
		}
		return val
	}
}

// ReadSecret prompts for passwords until it is validated
func readSecret(p *prompt.Prompt, label string, validate func(string) error) string {
	for {
		val, err := p.ReadPassword(label)
		if err != nil {
			continue
		}
		if validate != nil {
			if verr := validate(val); verr != nil {
				fmt.Println("Invalid input:", verr)
				continue
			}
		}
		return val
	}
}

func (h *Handler) Whoami(sess *session.Session) error {
	user, err := h.userService.GetByUsername(sess.Username)
	if err != nil {
		return err
	}

	fmt.Println(style.TitleStyle.Render("\nAccount Details"))
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
