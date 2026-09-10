package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/cli"
)

func main() {
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
			cli.RegisterUser(reader)

		case "login":
			fmt.Println("Login Screen")
			cli.LoginUser(reader)

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
