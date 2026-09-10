package service

import (
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/auth"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/models"
	"github.com/Shiwang0-0/Containerized-CLI-Login-System/internals/repository"
)

type UserService struct {
	repository *repository.UserRepository
}

func NewUserService(repository *repository.UserRepository) *UserService {
	return &UserService{
		repository: repository,
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

	if err := auth.CheckPassword(user.PasswordHash, password); err != nil {
		return models.User{}, models.ErrInvalidCredentials
	}

	return user, nil
}
