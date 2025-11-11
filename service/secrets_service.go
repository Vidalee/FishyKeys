package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"

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

func (s *SecretsService) ListSecrets(ctx context.Context) ([]*gensecrets.SecretInfoSummary, error) {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	secrets, err := s.secretsRepository.ListSecretsForUser(ctx, jwtClaims.UserID)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error listing secrets: %w", err))
	}

	// Retrieve all roles and users to build maps for easy lookup when listing access for each secret
	roles, err := s.rolesRepository.ListRoles(ctx)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving roles: %w", err))
	}
	users, err := s.usersRepository.ListUsers(ctx)
	if err != nil {
		return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving users: %w", err))
	}
	roleMap := make(map[int]gensecrets.Role)
	for _, role := range roles {
		roleMap[role.ID] = gensecrets.Role{
			ID:        role.ID,
			Name:      role.Name,
			Color:     role.Color,
			Admin:     role.Admin,
			CreatedAt: role.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: role.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	userMap := make(map[int]gensecrets.User)
	for _, user := range users {
		userMap[user.ID] = gensecrets.User{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	secretInfos := make([]*gensecrets.SecretInfoSummary, 0, len(secrets))
	for _, secret := range secrets {
		owner, err := s.usersRepository.GetUserByID(ctx, secret.OwnerUserId)
		if err != nil {
			return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving secret owner: %w", err))
		}

		secretUserIds, roleUserIds, err := s.secretsAccessRepository.GetAccessesBySecretPath(ctx, secret.Path)
		if err != nil {
			return nil, gensecrets.MakeInternalError(fmt.Errorf("error retrieving secret accesses for secret %s: %w", secret.Path, err))
		}

		var secretRoles []*gensecrets.Role
		for _, roleID := range roleUserIds {
			if role, exists := roleMap[roleID]; exists {
				secretRoles = append(secretRoles, &role)
			}
		}
		var secretUsers []*gensecrets.User
		for _, userID := range secretUserIds {
			if user, exists := userMap[userID]; exists {
				secretUsers = append(secretUsers, &user)
			}
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
			Roles:     secretRoles,
			Users:     secretUsers,
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

func (s *SecretsService) GetSecret(ctx context.Context, payload *gensecrets.GetSecretPayload) (*gensecrets.SecretInfo, error) {
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
	var authorizedRolesPayload []*gensecrets.Role
	for _, role := range authorizedRoles {
		authorizedRolesPayload = append(authorizedRolesPayload, &gensecrets.Role{
			ID:        role.ID,
			Name:      role.Name,
			Color:     role.Color,
			Admin:     role.Admin,
			CreatedAt: role.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: role.UpdatedAt.Format("2006-01-02T15:04:05Z"),
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

func (s *SecretsService) GetSecretValue(ctx context.Context, payload *gensecrets.GetSecretValuePayload) (res *gensecrets.GetSecretValueResult, err error) {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	decryptedValue, decodedPath, err := s.getSecretValue(ctx, jwtClaims.UserID, payload.Path)
	if err != nil {
		return nil, err
	}

	return &gensecrets.GetSecretValueResult{
		Value: &decryptedValue,
		Path:  &decodedPath,
	}, nil
}

// grpcurl -plaintext -H metadata:token -d '{"path": "/folder/secret"}' 172.19.32.1:8090 secrets.Secrets/OperatorGetSecretValue
func (s *SecretsService) OperatorGetSecretValue(ctx context.Context, payload *gensecrets.OperatorGetSecretValuePayload) (*gensecrets.OperatorGetSecretValueResult, error) {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	decryptedValue, decodedPath, err := s.getSecretValue(ctx, jwtClaims.UserID, payload.Path)
	if err != nil {
		return nil, err
	}

	return &gensecrets.OperatorGetSecretValueResult{
		Value: &decryptedValue,
		Path:  &decodedPath,
	}, nil
}

func (s *SecretsService) getSecretValue(ctx context.Context, userID int, path string) (string, string, error) {
	roleIDs, err := s.userRolesRepository.GetUserRoleIDs(ctx, userID)
	if err != nil {
		return "", "", gensecrets.MakeInternalError(fmt.Errorf("error retrieving user roles: %w", err))
	}

	decodedPath, err := base64.StdEncoding.DecodeString(path)
	if err != nil {
		return "", "", gensecrets.MakeInvalidParameters(fmt.Errorf("invalid path encoding: %w", err))
	}
	decodedPathStr := string(decodedPath)

	hasAccess, err := s.secretsRepository.HasAccess(ctx, decodedPathStr, &userID, roleIDs)
	if err != nil {
		return "", "", gensecrets.MakeInternalError(fmt.Errorf("error checking access to secret: %w", err))
	}

	if !hasAccess {
		return "", "", gensecrets.MakeForbidden(fmt.Errorf("you do not have access to this secret"))
	}

	decryptedSecret, err := s.secretsRepository.GetSecretByPath(ctx, s.keyManager, decodedPathStr)
	if err != nil {
		if errors.Is(err, repository.ErrSecretNotFound) {
			return "", "", gensecrets.MakeSecretNotFound(fmt.Errorf("secret not found at path: %s", decodedPathStr))
		}
		return "", "", gensecrets.MakeInternalError(fmt.Errorf("error retrieving secret: %w", err))
	}

	return decryptedSecret.DecryptedValue, decodedPathStr, nil
}

func (s *SecretsService) UpdateSecret(ctx context.Context, payload *gensecrets.UpdateSecretPayload) error {
	// Guaranteed by the Authentified interceptor
	jwtClaims := ctx.Value("token").(*JwtClaims)

	decodedPath, err := base64.StdEncoding.DecodeString(payload.Path)
	if err != nil {
		return gensecrets.MakeInvalidParameters(fmt.Errorf("invalid path encoding: %w", err))
	}
	decodedPathStr := string(decodedPath)

	secrets, err := s.secretsRepository.ListSecretsForUser(ctx, jwtClaims.UserID)
	if err != nil {
		return gensecrets.MakeInternalError(fmt.Errorf("error listing secrets: %w", err))
	}

	var secretToUpdate *repository.Secret
	for i := range secrets {
		if secrets[i].Path == decodedPathStr {
			secretToUpdate = &secrets[i]
			break
		}
	}
	if secretToUpdate == nil {
		return gensecrets.MakeForbidden(fmt.Errorf("you do not have access to this secret"))
	}

	userRoles, err := s.userRolesRepository.GetUserRoleIDs(ctx, jwtClaims.UserID)
	if secretToUpdate.OwnerUserId != jwtClaims.UserID || !slices.Contains(userRoles, 1) {
		return gensecrets.MakeForbidden(fmt.Errorf("only the owner or an admin can update the secret"))
	}

	if !slices.Contains(payload.AuthorizedRoles, 1) {
		return gensecrets.MakeInvalidParameters(fmt.Errorf("admin role (ID 1) must always have access to the secret"))
	}

	err = s.secretsRepository.UpdateSecret(ctx, s.keyManager, decodedPathStr, payload.Value)
	if err != nil {
		return gensecrets.MakeInternalError(fmt.Errorf("error updating secret: %w", err))
	}

	authorizedUsersIds, authorizedRolesIds, err := s.secretsAccessRepository.GetAccessesBySecretPath(ctx, decodedPathStr)
	if err != nil {
		return gensecrets.MakeInternalError(fmt.Errorf("error retrieving current secret accesses: %w", err))
	}

	for _, newUserId := range payload.AuthorizedUsers {
		found := slices.Contains(authorizedUsersIds, newUserId)
		if !found {
			err = s.secretsAccessRepository.GrantUsersAccess(ctx, decodedPathStr, []int{newUserId})
			if err != nil {
				return gensecrets.MakeInternalError(fmt.Errorf("error adding authorized user %d: %w", newUserId, err))
			}
		}
	}
	for _, existingUserId := range authorizedUsersIds {
		found := slices.Contains(payload.AuthorizedUsers, existingUserId)
		if !found {
			err = s.secretsAccessRepository.RevokeUserAccess(ctx, decodedPathStr, existingUserId)
			if err != nil {
				return gensecrets.MakeInternalError(fmt.Errorf("error removing authorized user %d: %w", existingUserId, err))
			}
		}
	}

	for _, newRoleId := range payload.AuthorizedRoles {
		found := slices.Contains(authorizedRolesIds, newRoleId)
		if !found {
			err = s.secretsAccessRepository.GrantRolesAccess(ctx, decodedPathStr, []int{newRoleId})
			if err != nil {
				return gensecrets.MakeInternalError(fmt.Errorf("error adding authorized role %d: %w", newRoleId, err))
			}
		}
	}
	for _, existingRoleId := range authorizedRolesIds {
		found := slices.Contains(payload.AuthorizedRoles, existingRoleId)
		if !found {
			err = s.secretsAccessRepository.RevokeRoleAccess(ctx, decodedPathStr, existingRoleId)
			if err != nil {
				return gensecrets.MakeInternalError(fmt.Errorf("error removing authorized role %d: %w", existingRoleId, err))
			}
		}
	}

	return nil
}
