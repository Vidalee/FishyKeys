package service

import (
	"context"
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

func (s *SecretsService) GetSecretValue(context.Context, *gensecrets.GetSecretValuePayload) (res *gensecrets.GetSecretValueResult, err error) {
	return nil, nil
}

func (s *SecretsService) GetSecret(context.Context, *gensecrets.GetSecretPayload) (res *gensecrets.SecretInfo, err error) {
	return nil, nil
}

func (s *SecretsService) CreateSecret(ctx context.Context, payload *gensecrets.CreateSecretPayload) error {
	return nil
}
