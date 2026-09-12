package cli_handler

import (
	"fmt"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/session"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
)

func (h *Handler) Login(p *prompt.Prompt, timeout time.Duration) (*session.Session, error) {
	username, lastLogin, err := h.LoginUser(p)
	if err != nil {
		return nil, err
	}

	tokenStr, expiresAt, err := h.jwtService.Generate(username, timeout)
	if err != nil {
		return nil, fmt.Errorf("issuing session token: %w", err)
	}

	if err := h.vault.SaveToken(tokenStr); err != nil {
		return nil, fmt.Errorf("persisting session: %w", err)
	}

	sess := session.New(username, timeout, lastLogin)
	sess.ExpiresAt = expiresAt // keep in sync with the token's actual exp
	return sess, nil
}

func (h *Handler) LoginUser(p *prompt.Prompt) (username string, lastLogin *time.Time, err error) {

	username, err = readLine(p, "Username: ", auth.ValidateUsername)
	if err != nil {
		return "", nil, err
	}

	password, err := readSecret(p, "Password: ", auth.ValidatePassword)
	if err != nil {
		return "", nil, err
	}

	user, err := h.userService.Authenticate(username, password)
	if err != nil {
		return "", nil, err
	}

	// once username and password are valid, ask for TOTP and verify it
	if user.TOTPEnabled {
		fmt.Print(style.PromptStyle.Render("Authenticator code: "))

		code, err := readLine(p, "Authenticator code: ", nil)
		if err != nil {
			return "", nil, err
		}
		if err := h.userService.VerifyTOTP(username, code); err != nil {
			fmt.Println(style.ErrorStyle.Render("Login failed:"), err)
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

	fmt.Println(style.SuccessStyle.Render("Login successful"))

	// create session on successfull login
	return username, prevLogin, nil
}

func (h *Handler) RefreshSession(sess *session.Session, timeout time.Duration) {
	sess.Refresh(timeout)
	if tokenStr, expiresAt, err := h.jwtService.Generate(sess.Username, timeout); err == nil {
		sess.ExpiresAt = expiresAt
		_ = h.vault.SaveToken(tokenStr)
	}
}

// is there an token present already , if yes return that session of the user
func (h *Handler) TryResume() *session.Session {
	tokenStr, err := h.vault.LoadToken()
	if err != nil || tokenStr == "" {
		return nil
	}
	claims, err := h.jwtService.Verify(tokenStr)
	if err != nil {
		_ = h.vault.ClearToken()
		return nil
	}
	return &session.Session{
		Username:   claims.Username,
		LoggedInAt: claims.IssuedAt.Time,
		ExpiresAt:  claims.ExpiresAt.Time,
	}
}

func (h *Handler) Logout() {
	_ = h.vault.ClearToken()
}
