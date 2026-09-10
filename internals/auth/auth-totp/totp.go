package totp

import (
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type Service struct {
	issuer string
}

func NewService(issuer string) *Service {
	return &Service{
		issuer: issuer,
	}
}

func (s *Service) GenerateSecret(username string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: username,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
}

func (s *Service) Validate(secret, code string) bool {
	return totp.Validate(code, secret)
}
