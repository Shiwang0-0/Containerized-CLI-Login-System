// internals/auth/register.go
package auth

import "strings"

type RegisterRequest struct {
	Credentials
}

func NewRegisterRequest(username, password string) (RegisterRequest, error) {
	req := RegisterRequest{Credentials{
		Username: strings.TrimSpace(username),
		Password: password,
	}}
	if err := req.Validate(); err != nil {
		return RegisterRequest{}, err
	}
	return req, nil
}

func (r RegisterRequest) Validate() error {
	if err := ValidateUsername(r.Username); err != nil {
		return err
	}
	return ValidatePassword(r.Password)
}
