package service

import (
	"errors"
	"fmt"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	totp "github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth/auth-totp"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/models"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/repository"
)

type UserService struct {
	repository *repository.UserRepository
	totp       *totp.Service
}

func NewUserService(repository *repository.UserRepository, totpService *totp.Service) *UserService {
	return &UserService{
		repository: repository,
		totp:       totpService,
	}
}

func (s *UserService) Register(username, password string) (models.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return models.User{}, err
	}

	user := models.User{
		Username:     username,
		PasswordHash: hash,
		TOTPEnabled:  false,
	}

	if err := s.repository.Create(user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *UserService) Authenticate(username, password string) (models.User, error) {
	user, err := s.repository.FindByUsername(username)
	if err != nil {
		return models.User{}, models.ErrInvalidCredentials
	}

	err = auth.CheckPassword(user.PasswordHash, password)

	if err != nil {
		return models.User{}, models.ErrInvalidCredentials
	}

	return user, nil
}

func (s *UserService) StartTOTPSetup(username string) (string, string, error) {
	user, err := s.repository.FindByUsername(username)
	if err != nil {
		return "", "", err
	}
	if user.TOTPEnabled {
		return "", "", errors.New("2FA already enabled, disable it first to re-enroll")
	}

	key, err := s.totp.GenerateSecret(user.Username)
	if err != nil {
		return "", "", err
	}

	secret := key.Secret()
	if err := s.repository.UpdateTOTP(user.Username, secret, false); err != nil {
		return "", "", err
	}
	return secret, key.URL(), nil // secret for manual entry, URL for QR
}

func (s *UserService) ConfirmTOTPSetup(username string, code string) error {

	user, err := s.repository.FindByUsername(username)
	if err != nil {
		return err
	}

	fmt.Println(user.TOTPSecret)

	if user.TOTPSecret == "" {
		return errors.New("TOTP setup has not been started")
	}

	if !s.totp.Validate(user.TOTPSecret, code) {
		return errors.New("invalid authentication code")
	}

	return s.repository.UpdateTOTP(user.Username, user.TOTPSecret, true)
}

func (s *UserService) VerifyTOTP(username string, code string) error {

	user, err := s.repository.FindByUsername(username)
	if err != nil {
		return errors.New("invalid authentication attempt")
	}

	if !user.TOTPEnabled {
		return nil
	}

	if !s.totp.Validate(user.TOTPSecret, code) {
		return errors.New("invalid authentication code")
	}

	return nil
}

func (s *UserService) DisableTOTP(username, password string) error {
	user, err := s.repository.FindByUsername(username)
	if err != nil {
		return models.ErrInvalidCredentials
	}

	// Re-verify password before allowing a security-sensitive change
	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return models.ErrInvalidCredentials
	}

	if !user.TOTPEnabled {
		return errors.New("2FA is not enabled")
	}

	return s.repository.UpdateTOTP(username, "", false)
}
