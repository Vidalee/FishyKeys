package service

import (
	"context"
	"errors"
	"fmt"
	genusers "github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/repository"
	"golang.org/x/crypto/bcrypt"
)

type UsersService struct {
	keyManager               *crypto.KeyManager
	usersRepository          repository.UsersRepository
	globalSettingsRepository repository.GlobalSettingsRepository
}

func NewUsersService(
	keyManager *crypto.KeyManager,
	usersRepository repository.UsersRepository,
	globalSettingsRepository repository.GlobalSettingsRepository,
) *UsersService {
	return &UsersService{
		keyManager:               keyManager,
		usersRepository:          usersRepository,
		globalSettingsRepository: globalSettingsRepository,
	}
}

func (s *UsersService) CreateUser(ctx context.Context, payload *genusers.CreateUserPayload) (*genusers.CreateUserResult, error) {
	if payload.Username == "" || payload.Password == "" {
		return nil, genusers.MakeInvalidParameters(fmt.Errorf("username and password must be provided"))
	}

	if len(payload.Password) < 8 {
		return nil, genusers.MakeInvalidParameters(fmt.Errorf("password must be at least 8 characters long"))
	}

	_, err := s.usersRepository.GetUserByUsername(ctx, payload.Username)
	if err == nil {
		return nil, genusers.MakeUsernameTaken(fmt.Errorf("username already exists"))
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, genusers.InternalError("could not check user existence")
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, genusers.InternalError("could not encrypt password: " + err.Error())
	}

	userId, err := s.usersRepository.CreateUser(ctx, payload.Username, string(encryptedPassword))
	if err != nil {
		return nil, genusers.InternalError("could not create user")
	}

	return &genusers.CreateUserResult{Username: &payload.Username, ID: &userId}, nil
}

func (s *UsersService) AuthUser(ctx context.Context, payload *genusers.AuthUserPayload) (*genusers.AuthUserResult, error) {
	if payload.Username == "" || payload.Password == "" {
		return nil, genusers.MakeInvalidParameters(fmt.Errorf("username and password must be provided"))
	}

	user, err := s.usersRepository.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, genusers.MakeUnauthorized(fmt.Errorf("invalid username or password"))
		}
		return nil, genusers.InternalError("could not retrieve user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		return nil, genusers.MakeUnauthorized(fmt.Errorf("invalid username or password"))
	}

	token := "fake-auth-token"

	return &genusers.AuthUserResult{
		Username: &user.Username,
		Token:    &token,
	}, nil
}

func (s *UsersService) ListUsers(ctx context.Context) ([]*genusers.User, error) {
	users, err := s.usersRepository.ListUsers(ctx)
	if err != nil {
		return nil, genusers.InternalError("could not list users")
	}

	result := make([]*genusers.User, 0, len(users))
	for _, u := range users {
		result = append(result, &genusers.User{
			Username:  u.Username,
			CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: u.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (s *UsersService) DeleteUser(ctx context.Context, payload *genusers.DeleteUserPayload) error {
	if payload.Username == "" {
		return genusers.MakeInvalidParameters(fmt.Errorf("username must be provided"))
	}

	err := s.usersRepository.DeleteUser(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return genusers.MakeUserNotFound(fmt.Errorf("user not found"))
		}
		return genusers.InternalError("could not delete user")
	}

	return nil
}
