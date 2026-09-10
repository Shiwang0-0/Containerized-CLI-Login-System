package cli_handler

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/mdp/qrterminal"
)

func (h *Handler) EnableTOTP(reader *bufio.Reader, username string) error {
	secret, url, err := h.userService.StartTOTPSetup(username)
	if err != nil {
		return err
	}

	fmt.Println("\nHow would you like to set up 2FA?")
	fmt.Println("  1) Scan QR code")
	fmt.Println("  2) Enter key manually")
	fmt.Print("Choice: ")

	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	switch choice {
	case "1":
		fmt.Println("\nScan this in your authenticator app:")
		qrterminal.GenerateHalfBlock(url, qrterminal.L, os.Stdout)
		fmt.Println("\n(If scanning fails, you can enter this key manually: " + secret + ")")

	default:
		fmt.Println("\nOpen your authenticator app and add a new account manually:")
		fmt.Println("  Account name:", username)
		fmt.Println("  Key (secret): ", secret)
		fmt.Println("  Type: Time-based, SHA1, 6 digits, 30s")
	}

	fmt.Print("\nEnter the 6-digit code to confirm: ")
	code, _ := reader.ReadString('\n')
	code = strings.TrimSpace(code)

	if err := h.userService.ConfirmTOTPSetup(username, code); err != nil {
		fmt.Println("2FA setup failed:", err)
		return err
	}

	fmt.Println("2FA enabled successfully.")
	return nil
}

func (h *Handler) DisableTOTP(reader *bufio.Reader, username string) error {
	password := readSecret("Confirm password to disable 2FA: ", auth.ValidatePassword)

	if err := h.userService.DisableTOTP(username, password); err != nil {
		fmt.Println("Failed to disable 2FA:", err)
		return err
	}

	fmt.Println("2FA disabled.")
	return nil
}
