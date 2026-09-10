package cli

import (
	"bufio"
	"fmt"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
)

func RegisterUser(reader *bufio.Reader) auth.RegisterRequest {
	fmt.Println("\n Register")

	username := readLine(reader, "Username: ", auth.ValidateUsername)
	password := readSecret("Password: ", auth.ValidatePassword)

	req, _ := auth.NewRegisterRequest(username, password)

	fmt.Println("Registration input is valid!")
	fmt.Printf("%s\n", req)

	return req
}
