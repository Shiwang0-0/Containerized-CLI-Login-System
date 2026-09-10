package cli

import (
	"bufio"
	"fmt"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
)

func LoginUser(reader *bufio.Reader) auth.LoginRequest {
	fmt.Println("\n Login")

	username := readLine(reader, "Username: ", auth.ValidateUsername)
	password := readSecret("Password: ", auth.ValidatePassword)

	req, _ := auth.NewLoginRequest(username, password)

	fmt.Println("Login input is valid!")
	fmt.Printf("%s\n", req)

	return req
}
