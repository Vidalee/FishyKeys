package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	genusers "github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type JwtClaims struct {
	Username string `json:"username"`
	UserID   int    `json:"userid"`
	jwt.RegisteredClaims
}

type UsersService struct {
	keyManager               *crypto.KeyManager
	usersRepository          repository.UsersRepository
	globalSettingsRepository repository.GlobalSettingsRepository
	secretsRepository        repository.SecretsRepository
}

func NewUsersService(
	keyManager *crypto.KeyManager,
	usersRepository repository.UsersRepository,
	globalSettingsRepository repository.GlobalSettingsRepository,
	secretsRepository repository.SecretsRepository,
) *UsersService {
	return &UsersService{
		keyManager:               keyManager,
		usersRepository:          usersRepository,
		globalSettingsRepository: globalSettingsRepository,
		secretsRepository:        secretsRepository,
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
		return nil, genusers.MakeInternalError(fmt.Errorf("could not check user existence: %w", err))
	}

	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not encrypt password: %w", err))
	}

	userId, err := s.usersRepository.CreateUser(ctx, payload.Username, string(encryptedPassword))
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not create user: %w", err))
	}

	return &genusers.CreateUserResult{Username: payload.Username, ID: userId}, nil
}

func (s *UsersService) AuthUser(ctx context.Context, payload *genusers.AuthUserPayload) (*genusers.AuthUserResult, error) {
	if payload.Username == "" || payload.Password == "" {
		return nil, genusers.MakeInvalidParameters(fmt.Errorf("username and password must be provided"))
	}

	if payload.Username == "system" {
		return nil, genusers.MakeUnauthorized(fmt.Errorf("system user cannot be authenticated"))
	}

	user, err := s.usersRepository.GetUserByUsername(ctx, payload.Username)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, genusers.MakeUnauthorized(fmt.Errorf("invalid username or password"))
		}
		return nil, genusers.MakeInternalError(fmt.Errorf("could not retrieve user: %w", err))
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		return nil, genusers.MakeUnauthorized(fmt.Errorf("invalid username or password"))
	}

	decryptedSecret, err := s.secretsRepository.GetSecretByPath(ctx, s.keyManager, "internal/jwt_signing_key")
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not retrieve JWT signing key: %w", err))
	}
	decodedSecret, err := base64.StdEncoding.DecodeString(decryptedSecret.DecryptedValue)
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not decode JWT signing key: %w", err))
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "FishyKeys",
		"sub":      user.Username,
		"iat":      jwt.NewNumericDate(time.Now()),
		"username": user.Username,
		"exp":      jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		"userid":   user.ID,
	})

	tokenString, err := token.SignedString(decodedSecret)
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not sign JWT token: %w", err))
	}

	return &genusers.AuthUserResult{
		Username: &user.Username,
		Token:    &tokenString,
	}, nil
}

func (s *UsersService) ListUsers(ctx context.Context) ([]*genusers.User, error) {
	users, err := s.usersRepository.ListUsers(ctx)
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not list users: %w", err))
	}

	result := make([]*genusers.User, 0, len(users))
	for _, u := range users {
		result = append(result, &genusers.User{
			ID:        u.ID,
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
		return genusers.MakeInternalError(fmt.Errorf("could not delete user: %w", err))
	}

	return nil
}
