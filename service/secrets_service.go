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
	secretsAccessRepository  repository.SecretsAccessRepository
}

func NewSecretsService(
	keyManager *crypto.KeyManager,
	usersRepository repository.UsersRepository,
	rolesRepository repository.RolesRepository,
	userRolesRepository repository.UserRolesRepository,
	globalSettingsRepository repository.GlobalSettingsRepository,
	secretsRepository repository.SecretsRepository,
	secretsAccessRepository repository.SecretsAccessRepository,
) *SecretsService {
	return &SecretsService{
		keyManager:               keyManager,
		usersRepository:          usersRepository,
		rolesRepository:          rolesRepository,
		userRolesRepository:      userRolesRepository,
		globalSettingsRepository: globalSettingsRepository,
		secretsRepository:        secretsRepository,
		secretsAccessRepository:  secretsAccessRepository,
	}
}

func (s *SecretsService) ListSecrets(ctx context.Context) (res []*gensecrets.SecretInfoSummary, err error) {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	secrets, err := s.secretsRepository.ListSecretsForUser(ctx, jwtClaims.UserID)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error listing secrets: %w", err))
	}

	secretInfos := make([]*gensecrets.SecretInfoSummary, 0, len(secrets))
	for _, secret := range secrets {
		owner, err := s.usersRepository.GetUserByID(ctx, secret.OwnerUserId)
		if err != nil {
			return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving secret owner: %w", err))
		}

		secretInfos = append(secretInfos, &gensecrets.SecretInfoSummary{
			Path: secret.Path,
			Owner: &gensecrets.User{
				ID:        owner.ID,
				Username:  owner.Username,
				CreatedAt: owner.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: owner.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			},
			CreatedAt: secret.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: secret.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	return secretInfos, nil
}

func (s *SecretsService) CreateSecret(ctx context.Context, payload *gensecrets.CreateSecretPayload) error {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	decodedPath, err := base64.StdEncoding.DecodeString(payload.Path)
	if err != nil {
		return gensecrets.MakeInvalidParameters(fmt.Errorf("invalid path encoding: %w", err))
	}
	decodedPathStr := string(decodedPath)

	if len(decodedPathStr) == 0 || decodedPathStr[0] != '/' {
		return gensecrets.MakeInvalidParameters(fmt.Errorf("path must start with '/'"))
	}

	_, err = s.secretsRepository.GetSecretByPath(ctx, s.keyManager, decodedPathStr)
	if err == nil {
		return gensecrets.MakeInvalidParameters(fmt.Errorf("secret already exists at path: %s", decodedPathStr))
	}
	if !errors.Is(err, repository.ErrSecretNotFound) {
		return gensecrets.MakeInternalError(fmt.Errorf("error checking if secret already exists: %w", err))
	}

	_, err = s.secretsRepository.CreateSecret(ctx, s.keyManager, decodedPathStr, jwtClaims.UserID, payload.Value)
	if err != nil {
		return gensecrets.MakeInternalError(fmt.Errorf("error creating secret: %w", err))
	}

	err = s.secretsAccessRepository.GrantUsersAccess(ctx, decodedPathStr, payload.AuthorizedUsers)
	if err != nil {
		return gensecrets.MakeInternalError(fmt.Errorf("error adding authorized users: %w", err))
	}

	err = s.secretsAccessRepository.GrantRolesAccess(ctx, decodedPathStr, payload.AuthorizedRoles)
	if err != nil {
		return gensecrets.MakeInternalError(fmt.Errorf("error adding authorized roles: %w", err))
	}

	return nil
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

	owner, err := s.usersRepository.GetUserByID(ctx, decryptedSecret.OwnerUserId)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving secret owner: %w", err))
	}

	authorizedUsers, err := s.usersRepository.GetUsersByIDs(ctx, decryptedSecret.AuthorizedUserIDs)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving authorized users: %w", err))
	}
	var authorizedUsersPayload []*gensecrets.User
	for _, user := range authorizedUsers {
		authorizedUsersPayload = append(authorizedUsersPayload, &gensecrets.User{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	authorizedRoles, err := s.rolesRepository.GetRolesByIDs(ctx, decryptedSecret.AuthorizedRoleIDs)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving authorized roles: %w", err))
	}
	var authorizedRolesPayload []*gensecrets.RoleType
	for _, role := range authorizedRoles {
		authorizedRolesPayload = append(authorizedRolesPayload, &gensecrets.RoleType{
			ID:   role.ID,
			Name: role.Name,
		})
	}

	return &gensecrets.SecretInfo{
		Path: decryptedSecret.Path,
		Owner: &gensecrets.User{
			ID:        owner.ID,
			Username:  owner.Username,
			CreatedAt: owner.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: owner.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
		AuthorizedUsers: authorizedUsersPayload,
		AuthorizedRoles: authorizedRolesPayload,
		CreatedAt:       decryptedSecret.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       decryptedSecret.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}, nil
}
