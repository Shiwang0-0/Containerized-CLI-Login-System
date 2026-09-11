package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(sql *sql.DB) *UserRepository {
	return &UserRepository{
		db: sql,
	}
}

func (r *UserRepository) Create(user models.User) error {
	_, err := r.db.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		user.Username, user.PasswordHash, // created at time is automatically set by the database
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) FindByUsername(username string) (models.User, error) {
	var u models.User
	query := `SELECT  username, password_hash, created_at,
	            totp_secret, totp_enabled,
	            failed_attempts, locked_until, lockout_level, last_lockout_at,
	            totp_failed_attempts, totp_locked_until, totp_lockout_level, totp_last_lockout_at, last_login_at 
	          	FROM users WHERE username = ?`

	err := r.db.QueryRow(query, username).Scan(
		&u.Username, &u.PasswordHash, &u.CreatedAt,
		&u.TOTPSecret, &u.TOTPEnabled,
		&u.FailedAttempts, &u.LockedUntil, &u.LockoutLevel, &u.LastLockoutAt,
		&u.TOTPFailedAttempts, &u.TOTPLockedUntil, &u.TOTPLockoutLevel, &u.TOTPLastLockoutAt, &u.LastLoginAt,
	)
	if err != nil {
		return models.User{}, err
	}
	return u, nil
}

func (r *UserRepository) UpdateTOTP(username string, secret string, enabled bool) error {
	result, err := r.db.Exec(` UPDATE users SET totp_secret = ?, totp_enabled = ? WHERE username = ?`, secret, enabled, username)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) IncrementFailedAttempts(username string) (int, error) {
	_, err := r.db.Exec(
		`UPDATE users SET failed_attempts = failed_attempts + 1 WHERE username = ?`, username)
	if err != nil {
		return 0, err
	}

	var count int
	err = r.db.QueryRow(`SELECT failed_attempts FROM users WHERE username = ?`, username).Scan(&count)
	return count, err
}

func (r *UserRepository) ResetFailedAttempts(username string) error {
	_, err := r.db.Exec(
		`UPDATE users SET failed_attempts = 0, lockout_level = 0, locked_until = NULL, last_lockout_at = NULL WHERE username = ?`,
		username,
	)
	return err
}

func (r *UserRepository) LockUser(username string, until time.Time, level int) error {
	_, err := r.db.Exec(
		`UPDATE users SET locked_until = ?, lockout_level = ?, last_lockout_at = ?, failed_attempts = 0 WHERE username = ?`, until, level, time.Now(), username,
	)
	return err
}

// TOTP equivalents for account lock

func (r *UserRepository) IncrementTOTPFailedAttempts(username string) (int, error) {
	_, err := r.db.Exec(
		`UPDATE users SET totp_failed_attempts = totp_failed_attempts + 1 WHERE username = ?`,
		username,
	)
	if err != nil {
		return 0, err
	}

	var count int
	err = r.db.QueryRow(`SELECT totp_failed_attempts FROM users WHERE username = ?`, username).Scan(&count)
	return count, err
}

func (r *UserRepository) ResetTOTPFailedAttempts(username string) error {
	_, err := r.db.Exec(
		`UPDATE users SET totp_failed_attempts = 0, totp_lockout_level = 0, totp_locked_until = NULL, totp_last_lockout_at = NULL WHERE username = ?`,
		username,
	)
	return err
}

func (r *UserRepository) LockTOTP(username string, until time.Time, level int) error {
	_, err := r.db.Exec(
		`UPDATE users SET totp_locked_until = ?, totp_lockout_level = ?, totp_last_lockout_at = ?, totp_failed_attempts = 0 WHERE username = ?`,
		until, level, time.Now(), username,
	)
	return err
}

func (r *UserRepository) UpdateLastLogin(username string, t time.Time) error {
	_, err := r.db.Exec(`UPDATE users SET last_login_at = ? WHERE username = ?`, t, username)
	return err
}
