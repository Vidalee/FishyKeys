package service

import (
	"context"
	"encoding/base64"
	gensecrets "github.com/Vidalee/FishyKeys/gen/secrets"
	"testing"

	"github.com/Vidalee/FishyKeys/internal/crypto"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSecretsTestService(t *testing.T) *SecretsService {
	keyManager := crypto.GetDefaultKeyManager()
	masterKey, err := crypto.GenerateSecret()
	require.NoError(t, err, "failed to generate master key")
	err = keyManager.SetNewMasterKey(masterKey, 1, 1)
	require.NoError(t, err, "failed to set new master key")
	usersRepo := repository.NewUsersRepository(testDB)
	globalSettingsRepo := repository.NewGlobalSettingsRepository(testDB)
	secretsRepo := repository.NewSecretsRepository(testDB)
	roleRepo := repository.NewRolesRepository(testDB)
	usersRolesRepo := repository.NewUserRolesRepository(testDB)
	secretsAccessRepo := repository.NewSecretsAccessRepository(testDB)
	return NewSecretsService(keyManager, usersRepo, roleRepo, usersRolesRepo, globalSettingsRepo, secretsRepo, secretsAccessRepo)
}

func clearSecretsServiceTables(t *testing.T, ctx context.Context) {
	err := testutil.ClearTable(ctx, "global_settings")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "secrets")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "users")
	require.NoError(t, err)
}

func TestSecretsService_CreateSecret(t *testing.T) {
	service := setupSecretsTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		path              string
		value             string
		authorizedUserIDs []int
		authorizedRoleIDs []int

		doDuplicate       bool
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "correct parameters",
			path:              "/test/secret1",
			value:             "test_value_1",
			authorizedUserIDs: []int{1, 2},
			authorizedRoleIDs: []int{1, 2},
		},
		{
			name:              "secret already exists",
			path:              "/test/secret1",
			value:             "test_value_1",
			authorizedUserIDs: []int{1, 2},
			authorizedRoleIDs: []int{1, 2},
			doDuplicate:       true,
			expectedError:     true,
			expectedErrorText: "secret already exists at path: /test/secret1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearSecretsServiceTables(t, ctx)

			userID1, err := service.usersRepository.CreateUser(ctx, "user1", "password1")
			require.NoError(t, err, "failed to create user1")
			_, err = service.usersRepository.CreateUser(ctx, "user2", "password2")
			require.NoError(t, err, "failed to create user2")
			err = service.rolesRepository.CreateRole(ctx, "role1", "#FFFFFF")
			require.NoError(t, err, "failed to create role1")
			err = service.rolesRepository.CreateRole(ctx, "role2", "#000000")
			require.NoError(t, err, "failed to create role2")

			payload := &gensecrets.CreateSecretPayload{
				Path:              base64.StdEncoding.EncodeToString([]byte(tt.path)),
				Value:             tt.value,
				AuthorizedMembers: tt.authorizedUserIDs,
				AuthorizedRoles:   tt.authorizedRoleIDs,
			}

			ctx = context.WithValue(ctx, "token", &JwtClaims{
				UserID: userID1,
			})

			err = service.CreateSecret(ctx, payload)
			if tt.doDuplicate {
				err = service.CreateSecret(ctx, payload)
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)
				secret, err := service.secretsRepository.GetSecretByPath(ctx, service.keyManager, tt.path)
				assert.NoError(t, err)
				assert.NotNil(t, secret)
				assert.Equal(t, tt.path, secret.Path)
				assert.Equal(t, tt.value, secret.DecryptedValue)
				assert.ElementsMatch(t, tt.authorizedUserIDs, secret.AuthorizedUserIDs)
				assert.ElementsMatch(t, tt.authorizedRoleIDs, secret.AuthorizedRoleIDs)
				assert.Equal(t, userID1, secret.OwnerUserId)
			}
		})
	}
}
