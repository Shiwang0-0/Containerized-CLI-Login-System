package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	totp "github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth/auth-totp"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth/jwt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/cli_handler"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/models"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/repository"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/service"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/session"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/vault"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {
	if err := godotenv.Load(); err != nil {
		// .env is optional (for docker)
	}

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FKolkata",
		os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("MYSQL_DATABASE"),
	)

	var err error
	db, err = sql.Open("mysql", uri)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 5; i++ {
		if err := db.Ping(); err != nil {
			log.Printf("MySQL connection attempt %d/5 failed: %v", i, err)
			time.Sleep(2 * time.Second)
			continue
		}
		log.Println("MySQL connected")
		return
	}
	log.Fatal("Error connecting to MySQL after 5 attempts")
}

func main() {
	handler, sessionTimeout, p := mustBootstrap()
	defer p.Close()

	fmt.Println(style.TitleStyle.Render("\n\nWelcome"))
	fmt.Println(style.InfoStyle.Render(
		fmt.Sprintf("Session timeout configured: %s. Type 'help' to see available commands.", sessionTimeout),
	))

	var currentSession *session.Session
	if sess := handler.TryResume(); sess != nil {
		currentSession = sess
		fmt.Println(style.SuccessStyle.Render("\nWelcome back, " + sess.Username + "."))
	}

	for {
		promptLabel := "> "
		if currentSession != nil {
			p.SetPostLoginMode()
			promptLabel = fmt.Sprintf("(%s) > ", currentSession.Username)
		} else {
			p.SetPreLoginMode()
		}

		command, err := p.ReadLine(promptLabel)
		if err != nil {
			fmt.Println(style.InfoStyle.Render("\nExiting..."))
			return
		}
		if command == "" {
			continue
		}

		// Check idle expiry
		if currentSession != nil && currentSession.IsExpired() {
			fmt.Println(style.WarningStyle.Render("\nYour session has expired due to inactivity. Please log in again."))
			handler.Logout()
			currentSession = nil
			p.SetPreLoginMode()
			continue
		}

		if currentSession == nil {
			if command == "exit" {
				fmt.Println(style.WarningStyle.Render("Exiting..."))
				return
			}
			currentSession = handlePreLoginCommand(handler, p, command, sessionTimeout)
			continue
		}

		// any command after loginrefreshes the session
		handler.RefreshSession(currentSession, sessionTimeout)

		if command == "exit" {
			fmt.Println("Exiting...")
			return
		}
		currentSession = handlePostLoginCommand(handler, p, currentSession, command)
	}
}

func mustBootstrap() (*cli_handler.Handler, time.Duration, *prompt.Prompt) {
	userRepository := repository.NewUserRepository(db)
	totpService := totp.NewTOTPService(os.Getenv("TOTP_ISSUER"))
	userService := service.NewUserService(userRepository, totpService)

	v, err := vault.New()
	if err != nil {
		log.Fatal("Failed to init local vault:", err)
	}
	signingKey, err := v.LoadOrCreateSigningKey()
	if err != nil {
		log.Fatal("Failed to load signing key:", err)
	}
	jwtService := jwt.NewJWTService(signingKey)

	handler := cli_handler.NewHandler(userService, jwtService, v)

	p, err := prompt.New()
	if err != nil {
		log.Fatal("Failed to start CLI prompt:", err)
	}

	return handler, session.SessionTimeout(), p
}

// handlePreLoginCommand runs one pre-login command and returns the new
// session if the command started one (nil otherwise).
func handlePreLoginCommand(handler *cli_handler.Handler, p *prompt.Prompt, command string, sessionTimeout time.Duration) *session.Session {
	switch command {
	case "register":
		fmt.Println(style.LabelStyle.Render("\nRegister Screen\n"))
		if err := handler.RegisterUser(p); err != nil {
			printCommandError(err, "Registration cancelled.")
		}
		return nil

	case "login":
		fmt.Println(style.LabelStyle.Render("\nLogin Screen"))
		sess, err := handler.Login(p, sessionTimeout)
		if err != nil {
			printCommandError(err, "Login cancelled.")
			return nil
		}
		p.SetPostLoginMode()
		printSessionStarted(sess)
		return sess // a new session on successful login

	case "help":
		printPreLoginHelp()
		return nil

	default:
		printUnknownCommand(command)
		return nil
	}
}

// handlePostLoginCommand runs one post-login command and returns the
// (possibly now-nil, if the user logged out) current session.
func handlePostLoginCommand(handler *cli_handler.Handler, p *prompt.Prompt, currentSession *session.Session, command string) *session.Session {
	switch command {
	case "whoami":
		if err := handler.Whoami(currentSession); err != nil {
			fmt.Println(style.ErrorStyle.Render("Error: " + err.Error()))
		}

	case "enable-2fa":
		if err := handler.EnableTOTP(p, currentSession.Username); err != nil {
			printCommandError(err, "2FA setup cancelled.")
		}

	case "disable-2fa":
		if err := handler.DisableTOTP(p, currentSession.Username); err != nil {
			printCommandError(err, "2FA disable cancelled.")
		}

	case "logout":
		fmt.Printf("%s %s\n",
			style.SuccessStyle.Render("Logging out"),
			style.InfoStyle.Render(currentSession.Username+"."),
		)
		handler.Logout()
		p.SetPreLoginMode()
		return nil

	case "help":
		printPostLoginHelp()

	default:
		printUnknownCommand(command)
	}
	return currentSession
}

func printCommandError(err error, cancelMsg string) {
	if errors.Is(err, models.ErrAborted) {
		fmt.Println(style.WarningStyle.Render("\n" + cancelMsg))
		return
	}
	fmt.Println(style.ErrorStyle.Render("Error: " + err.Error()))
}

func printSessionStarted(sess *session.Session) {
	fmt.Println(style.InfoStyle.Render(fmt.Sprintf(
		"Session started. Expires at %s.\n",
		sess.ExpiresAt.Format(time.Kitchen),
	)))
}

func printUnknownCommand(command string) {
	fmt.Println(style.WarningStyle.Render(fmt.Sprintf("Unknown command: %q. Type 'help' for available commands.", command)))
}

func printPreLoginHelp() {
	fmt.Println(style.InfoStyle.Render("\nAvailable commands:"))
	fmt.Println("  register  - create a new user")
	fmt.Println("  login     - login with username/password")
	fmt.Println("  help      - show available commands")
	fmt.Println("  exit      - quit program")
	fmt.Println()
}

func printPostLoginHelp() {
	fmt.Println(style.InfoStyle.Render("\nAvailable commands:"))
	fmt.Println("  whoami       - show current user details")
	fmt.Println("  enable-2fa   - turn on TOTP-based 2FA")
	fmt.Println("  disable-2fa  - turn off TOTP-based 2FA")
	fmt.Println("  logout       - end session")
	fmt.Println("  help         - show available commands")
	fmt.Println("  exit         - quit program")
}
