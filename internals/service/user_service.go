package service

import (
	"errors"
	"fmt"
	"time"

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

	// Check password lockout before even checking the password, if lockout is valid just return invalid credentials no matter what
	if user.LockedUntil.Valid && time.Now().Before(user.LockedUntil.Time) {
		fmt.Println("didnt even checked")
		return models.User{}, models.ErrInvalidCredentials
	}

	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		count, incErr := s.repository.IncrementFailedAttempts(username)
		if incErr != nil {
			return models.User{}, models.ErrInvalidCredentials
		}

		if count >= auth.FailThreshold {
			// add lockout
			dur, level := auth.NextLockout(user.LastLockoutAt.Time, user.LockoutLevel)
			_ = s.repository.LockUser(username, time.Now().Add(dur), level)
		} else if count >= auth.WarnAfter {
			// give user a warning
			remaining := auth.FailThreshold - count
			return models.User{}, fmt.Errorf("%w (warning: %d attempts remaining before lockout)",
				models.ErrInvalidCredentials, remaining)
		}

		return models.User{}, models.ErrInvalidCredentials
	}

	// dont reset the counter for failed attempts yet, first make sure the 2FA passes (if any)
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

	// Check TOTP lockout, if there is lockout return a generic error, dont reveal password already succeeded.
	if user.TOTPLockedUntil.Valid && time.Now().Before(user.TOTPLockedUntil.Time) {
		return models.ErrInvalidCredentials
	}

	if !s.totp.Validate(user.TOTPSecret, code) {
		count, incErr := s.repository.IncrementTOTPFailedAttempts(username)
		if incErr != nil {
			return models.ErrInvalidCredentials
		}

		if count >= auth.FailThreshold {
			// failed attempt of TOTP
			dur, level := auth.NextLockout(user.TOTPLastLockoutAt.Time, user.TOTPLockoutLevel)
			_ = s.repository.LockTOTP(username, time.Now().Add(dur), level)
		} else if count >= auth.WarnAfter {
			// failed attempt of TOTP
			remaining := auth.FailThreshold - count
			return fmt.Errorf("%w (warning: %d attempts remaining before lockout)",
				models.ErrInvalidCredentials, remaining)
		}

		return models.ErrInvalidCredentials
	}

	// full login success reset BOTH counters.
	_ = s.repository.ResetFailedAttempts(username)
	_ = s.repository.ResetTOTPFailedAttempts(username)
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

func (s *UserService) ResetLoginAttempts(username string) error {
	return s.repository.ResetFailedAttempts(username)
}
