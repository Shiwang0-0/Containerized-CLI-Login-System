package repository

import (
	"database/sql"

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
	row := r.db.QueryRow(
		`SELECT username, password_hash, created_at FROM users WHERE username = ?`,
		username,
	)
	if err := row.Scan(&u.Username, &u.PasswordHash, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, models.ErrUserNotFound
		}
		return models.User{}, err
	}
	return u, nil
}
