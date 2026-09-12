package cli_handler

import (
	"fmt"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth/jwt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/models"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/service"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/session"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/vault"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
)

type Handler struct {
	userService *service.UserService
	jwtService  *jwt.JWTService
	vault       *vault.Vault
}

func NewHandler(userService *service.UserService, jwtService *jwt.JWTService, vault *vault.Vault) *Handler {
	return &Handler{userService: userService, jwtService: jwtService, vault: vault}
}

// ValidateFunc validates a single piece of input.
type ValidateFunc func(string) error

// ReadLine prompts on stdin until input passes validate.
func readLine(p *prompt.Prompt, label string, validate func(string) error) (string, error) {
	for {
		val, err := p.ReadLineSensitive(label)
		if err != nil {
			// Ctrl+C or Ctrl+D: cancel this flow
			return "", models.ErrAborted
		}
		if validate != nil {
			if verr := validate(val); verr != nil {
				fmt.Println("Invalid input:", verr)
				continue
			}
		}
		return val, nil
	}
}

// ReadSecret prompts for passwords until it is validated
func readSecret(p *prompt.Prompt, label string, validate func(string) error) (string, error) {
	for {
		val, err := p.ReadPassword(label)
		if err != nil {
			return "", models.ErrAborted
		}
		if validate != nil {
			if verr := validate(val); verr != nil {
				fmt.Println("Invalid input:", verr)
				continue
			}
		}
		return val, nil
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
