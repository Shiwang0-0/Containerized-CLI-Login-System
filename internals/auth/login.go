// internals/auth/login.go
package auth

import "strings"

type LoginRequest struct {
	Credentials
}

func NewLoginRequest(username, password string) (LoginRequest, error) {
	req := LoginRequest{Credentials{
		Username: strings.TrimSpace(username),
		Password: password,
	}}
	if err := req.Validate(); err != nil {
		return LoginRequest{}, err
	}
	return req, nil
}

func (r LoginRequest) Validate() error {
	if err := ValidateUsername(r.Username); err != nil {
		return err
	}
	return ValidatePassword(r.Password)
}
