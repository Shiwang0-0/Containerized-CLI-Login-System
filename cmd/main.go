package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	totp "github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth/auth-totp"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/cli_handler"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/prompt"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/repository"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/service"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/session"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/style"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {

	if err := godotenv.Load(); err != nil {
		// .env is optional (for docker)
	}

	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

	var err error
	db, err = sql.Open("mysql", uri)

	if err != nil {
		log.Fatal(err)
	}

	connected := false

	for i := 1; i <= 5; i++ {
		if err := db.Ping(); err != nil {
			log.Printf("MySQL connection attempt %d/5 failed: %v", i, err)
			time.Sleep(2 * time.Second)
		} else {
			log.Println("MySQL connected")
			connected = true
			break
		}
	}

	if !connected {
		log.Fatal("Error connecting to MySQL after 5 attempts")
	}

}

func main() {

	userRepository := repository.NewUserRepository(db)
	totpService := totp.NewService(os.Getenv("TOTP_ISSUER"))
	userService := service.NewUserService(userRepository, totpService)
	handler := cli_handler.NewHandler(userService)

	sessionTimeout := session.SessionTimeout()

	p, err := prompt.New()
	if err != nil {
		log.Fatal("Failed to start CLI prompt:", err)
	}
	defer p.Close()

	fmt.Println(style.TitleStyle.Render("\n\nWelcome"))
	fmt.Println(style.InfoStyle.Render(
		fmt.Sprintf("Session timeout: %s. Type 'help' to see available commands.", sessionTimeout),
	))

	var currentSession *session.Session

	for {

		if currentSession != nil && currentSession.IsExpired() {
			fmt.Println("\nYour session has expired due to inactivity. Please log in again.")
			currentSession = nil
			p.SetPreLoginMode()
		}

		var promptLabel string
		if currentSession == nil {
			p.SetPreLoginMode()
			promptLabel = "> "
		} else {
			p.SetPostLoginMode()
			promptLabel = fmt.Sprintf("(%s) > ", currentSession.Username)
		}

		command, err := p.ReadLine(promptLabel)
		if err != nil {
			if err == prompt.ErrInterrupted {
				fmt.Println(style.WarningStyle.Render(
					"\nUse 'exit' or press Ctrl+D to exit.",
				))
				continue
			}

			// io.EOF (Ctrl+D) or real error: exit cleanly
			fmt.Println(style.InfoStyle.Render("\nExiting..."))
			return
		}

		if command == "" {
			continue
		}

		// pre login commands
		if currentSession == nil {
			switch command {
			case "register":
				fmt.Println(style.LabelStyle.Render("\nRegister Screen\n"))
				if err := handler.RegisterUser(p); err != nil {
					fmt.Println(style.ErrorStyle.Render("Error:", err.Error()))
				}

			case "login":
				fmt.Println(style.LabelStyle.Render("\nLogin Screen\n"))
				username, lastLogin, err := handler.LoginUser(p)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				// as soon as login finished, create a new session
				currentSession = session.New(username, sessionTimeout, lastLogin)
				p.SetPostLoginMode() // after login, change the command mode
				fmt.Printf(style.InfoStyle.Render(fmt.Sprintf(
					"Session started. Expires at %s.\n",
					currentSession.ExpiresAt.Format(time.Kitchen),
				)))

			case "help":
				fmt.Println(style.InfoStyle.Render("\nAvailable commands:"))
				fmt.Println("  register  - create a new user")
				fmt.Println("  login     - login with username/password")
				fmt.Println("  help      - show available commands")
				fmt.Println("  exit      - quit program")
				fmt.Println()

			case "exit":
				fmt.Println(style.WarningStyle.Render("Exiting..."))
				return

			case "":
				continue

			default:
				fmt.Println(style.WarningStyle.Render(fmt.Sprintf("Unknown command: %q. Type 'help' for available commands.", command)))
			}
			continue
		}

		// post login commands
		// any validate command post login, refreshes the current session
		currentSession.Refresh(sessionTimeout)
		switch command {
		case "whoami":
			if err := handler.Whoami(currentSession); err != nil {
				fmt.Println("Error:", err)
			}
		case "enable-2fa":
			if err := handler.EnableTOTP(p, currentSession.Username); err != nil {
				fmt.Println("Error:", err)
			}

		case "disable-2fa":
			if err := handler.DisableTOTP(p, currentSession.Username); err != nil {
				fmt.Println("Error:", err)
			}

		case "logout":
			// session end (explicit)
			fmt.Printf("%s %s\n",
				style.SuccessStyle.Render("Logging out"),
				style.InfoStyle.Render(currentSession.Username+"."),
			)
			currentSession = nil
			p.SetPreLoginMode() // after log out change the mode

		case "help":
			fmt.Println(style.InfoStyle.Render("\nAvailable commands:"))
			fmt.Println("  whoami       - show current user details")
			fmt.Println("  enable-2fa   - turn on TOTP-based 2FA")
			fmt.Println("  disable-2fa  - turn off TOTP-based 2FA")
			fmt.Println("  logout       - end session")
			fmt.Println("  help         - show available commands")
			fmt.Println("  exit         - quit program")

		case "exit":
			fmt.Println("Exiting...")
			return

		case "":
			continue

		default:
			fmt.Println(style.WarningStyle.Render(fmt.Sprintf("Unknown command: %q. Type 'help' for available commands.", command)))
		}
	}
}
