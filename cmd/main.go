package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	totp "github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth/auth-totp"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/cli_handler"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/repository"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/service"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

var db *sql.DB

func init() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env")
	}

	dbUser := os.Getenv("MYSQL_USER")
	dbPassword := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DATABASE")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	uri := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

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

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome")
	fmt.Println("Type 'help' to see available commands.")

	var currentUser string // empty string means not loggedin

	for {
		if currentUser == "" {
			fmt.Print("> ")
		} else {
			fmt.Printf("(%s) > ", currentUser)
		}

		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		command := strings.TrimSpace(strings.ToLower(input))

		// pre login commands
		if currentUser == "" {
			switch command {
			case "register":
				fmt.Println("Register Screen")
				if err := handler.RegisterUser(reader); err != nil {
					fmt.Println("Error:", err)
				}

			case "login":
				fmt.Println("Login Screen")
				username, err := handler.LoginUser(reader)
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}
				currentUser = username

			case "help":
				fmt.Println("\nAvailable commands:")
				fmt.Println("  register  - create a new user")
				fmt.Println("  login     - login with username/password")
				fmt.Println("  help      - show available commands")
				fmt.Println("  exit      - quit program")
				fmt.Println()

			case "exit":
				fmt.Println("Exiting...")
				return

			case "":
				continue

			default:
				fmt.Printf("Unknown command: %q. Type 'help' for available commands.\n", command)
			}
			continue
		}

		// post login commands
		switch command {
		case "enable-2fa":
			if err := handler.EnableTOTP(reader, currentUser); err != nil {
				fmt.Println("Error:", err)
			}

		case "disable-2fa":
			if err := handler.DisableTOTP(reader, currentUser); err != nil {
				fmt.Println("Error:", err)
			}

		case "logout":
			fmt.Println("Logging out.")
			currentUser = ""

		case "help":
			fmt.Println("\nAvailable commands:")
			fmt.Println("  enable-2fa   - turn on TOTP-based 2FA")
			fmt.Println("  disable-2fa  - turn off TOTP-based 2FA")
			fmt.Println("  logout       - end session")
			fmt.Println("  exit         - quit program")

		case "exit":
			fmt.Println("Exiting...")
			return

		case "":
			continue

		default:
			fmt.Printf("Unknown command: %q. Type 'help' for available commands.\n", command)
		}
	}
}
