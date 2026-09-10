package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	userService := service.NewUserService(userRepository)
	handler := cli_handler.NewHandler(userService)

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome")
	fmt.Println("Type 'help' to see available commands.")

	for {
		fmt.Print("> ")

		input, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		command := strings.TrimSpace(strings.ToLower(input))

		switch command {
		case "register":
			fmt.Println("Register Screen")
			if err := handler.RegisterUser(reader); err != nil {
				fmt.Println("Error:", err)
			}

		case "login":
			fmt.Println("Login Screen")
			if err := handler.LoginUser(reader); err != nil {
				fmt.Println("Error:", err)
			}

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
	}
}
