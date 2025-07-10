package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	gensecrets "github.com/Vidalee/FishyKeys/gen/secrets"
	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/repository"
)

type SecretsService struct {
	keyManager               *crypto.KeyManager
	usersRepository          repository.UsersRepository
	rolesRepository          repository.RolesRepository
	userRolesRepository      repository.UserRolesRepository
	globalSettingsRepository repository.GlobalSettingsRepository
	secretsRepository        repository.SecretsRepository
}

func NewSecretsService(
	keyManager *crypto.KeyManager,
	usersRepository repository.UsersRepository,
	rolesRepository repository.RolesRepository,
	userRolesRepository repository.UserRolesRepository,
	globalSettingsRepository repository.GlobalSettingsRepository,
	secretsRepository repository.SecretsRepository,
) *SecretsService {
	return &SecretsService{
		keyManager:               keyManager,
		usersRepository:          usersRepository,
		rolesRepository:          rolesRepository,
		userRolesRepository:      userRolesRepository,
		globalSettingsRepository: globalSettingsRepository,
		secretsRepository:        secretsRepository,
	}
}

func (s *SecretsService) GetSecretValue(ctx context.Context, payload *gensecrets.GetSecretValuePayload) (res *gensecrets.GetSecretValueResult, err error) {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	roleIDs, err := s.userRolesRepository.GetUserRoleIDs(ctx, jwtClaims.UserID)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving user roles: %w", err))
	}

	decodedPath, err := base64.StdEncoding.DecodeString(payload.Path)
	if err != nil {
		return nil, gensecrets.MakeInvalidParameters(fmt.Errorf("invalid path encoding: %w", err))
	}
	decodedPathStr := string(decodedPath)

	hasAccess, err := s.secretsRepository.HasAccess(ctx, decodedPathStr, &jwtClaims.UserID, roleIDs)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error checking access to secret: %w", err))
	}

	if !hasAccess {
		return nil, gensecrets.MakeForbidden(fmt.Errorf("you do not have access to this secret"))
	}

	decryptedSecret, err := s.secretsRepository.GetSecretByPath(ctx, s.keyManager, decodedPathStr)
	if err != nil {
		if errors.Is(err, repository.ErrSecretNotFound) {
			return nil, gensecrets.MakeSecretNotFound(fmt.Errorf("secret not found at path: %s", decodedPathStr))
		}
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving secret: %w", err))
	}

	return &gensecrets.GetSecretValueResult{
		Value: &decryptedSecret.DecryptedValue,
		Path:  &decryptedSecret.Path,
	}, nil
}

func (s *SecretsService) GetSecret(ctx context.Context, payload *gensecrets.GetSecretPayload) (res *gensecrets.SecretInfo, err error) {
	return nil, nil
}

func (s *SecretsService) CreateSecret(ctx context.Context, payload *gensecrets.CreateSecretPayload) error {
	return nil
}
