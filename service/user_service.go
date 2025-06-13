package service

import (
	"context"
	"errors"
	"log"

	genusers "github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users repository.UsersRepository
}

func NewUserService(users repository.UsersRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Create(ctx context.Context, payload *genusers.CreatePayload) (*genusers.CreateResult, error) {
	if payload.Username == "" || payload.Password == "" {
		return nil, genusers.InvalidInput("username and password must be provided")
	}

	_, err := s.users.GetUserByUsername(ctx, payload.Username)
	if err == nil {
		return nil, genusers.UsernameTaken("username already exists")
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		log.Printf("error checking user: %v", err)
		return nil, genusers.InternalError("could not check user existence")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("error hashing password: %v", err)
		return nil, genusers.InternalError("could not hash password")
	}

	err = s.users.CreateUser(ctx, payload.Username, string(hashedPassword))
	if err != nil {
		log.Printf("error creating user: %v", err)
		return nil, genusers.InternalError("could not create user")
	}

	return &genusers.CreateResult{Username: &payload.Username}, nil
}

func (s *UserService) Auth(ctx context.Context, payload *genusers.AuthPayload) (*genusers.AuthResult, error) {
	user, err := s.users.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, genusers.Unauthorized("invalid username or password")
		}
		log.Printf("error retrieving user for auth: %v", err)
		return nil, genusers.InternalError("could not authenticate user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		return nil, genusers.Unauthorized("invalid username or password")
	}

	token := "fake-auth-token"

	return &genusers.AuthResult{
		Username: &user.Username,
		Token:    &token,
	}, nil
}

func (s *UserService) List(ctx context.Context) ([]*genusers.User, error) {
	users, err := s.users.ListUsers(ctx)
	if err != nil {
		log.Printf("error listing users: %v", err)
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

func (s *UserService) Delete(ctx context.Context, payload *genusers.DeletePayload) error {
	if payload.Username == "" {
		return genusers.InvalidInput("username must be provided")
	}

	err := s.users.DeleteUser(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return genusers.UserNotFound("user does not exist")
		}
		log.Printf("error deleting user: %v", err)
		return genusers.InternalError("could not delete user")
	}

	return nil
}
