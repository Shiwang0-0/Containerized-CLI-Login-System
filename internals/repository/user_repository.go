package repository

import (
	"database/sql"
	"fmt"

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
		// TODO: detect duplicate-key error specifically and return auth.ErrUsernameTaken
		return err
	}
	return nil
}

func (r *UserRepository) FindByUsername(username string) (models.User, error) {
	var u models.User

	row := r.db.QueryRow(`
		SELECT username, password_hash, created_at, totp_secret, totp_enabled FROM users WHERE username = ?`, username)

	if err := row.Scan(&u.Username, &u.PasswordHash, &u.CreatedAt, &u.TOTPSecret, &u.TOTPEnabled); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, models.ErrUserNotFound
		}

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
