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
	err = testutil.ClearTable(ctx, "roles")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "user_roles")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "secrets_access")
	require.NoError(t, err)
}

func TestSecretsService_CreateSecret(t *testing.T) {
	service := setupSecretsTestService(t)
	ctx := context.Background()

	tests := []struct {
		name                    string
		path                    string
		value                   string
		addCreatedUsersAndRoles bool
		doDuplicate             bool
		expectedError           bool
		expectedErrorText       string
	}{
		{
			name:                    "correct parameters",
			path:                    "/test/secret1",
			value:                   "test_value_1",
			addCreatedUsersAndRoles: true,
		},
		{
			name:                    "secret already exists",
			path:                    "/test/secret1",
			value:                   "test_value_1",
			addCreatedUsersAndRoles: true,
			doDuplicate:             true,
			expectedError:           true,
			expectedErrorText:       "secret already exists at path: /test/secret1",
		},
		{
			name:                    "secret already exists",
			path:                    "test/secret1",
			value:                   "test_value_1",
			addCreatedUsersAndRoles: true,
			expectedError:           true,
			expectedErrorText:       "path must start with '/'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearSecretsServiceTables(t, ctx)

			userID1, err := service.usersRepository.CreateUser(ctx, "user1", "password1")
			require.NoError(t, err, "failed to create user1")
			userID2, err := service.usersRepository.CreateUser(ctx, "user2", "password2")
			require.NoError(t, err, "failed to create user2")
			roleID1, err := service.rolesRepository.CreateRole(ctx, "role1", "#FFFFFF")
			require.NoError(t, err, "failed to create role1")
			roleID2, err := service.rolesRepository.CreateRole(ctx, "role2", "#000000")
			require.NoError(t, err, "failed to create role2")

			var roles []int
			var users []int
			if tt.addCreatedUsersAndRoles {
				roles = []int{roleID1, roleID2}
				users = []int{userID1, userID2}
			}

			payload := &gensecrets.CreateSecretPayload{
				Path:            base64.StdEncoding.EncodeToString([]byte(tt.path)),
				Value:           tt.value,
				AuthorizedUsers: users,
				AuthorizedRoles: roles,
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
				assert.ElementsMatch(t, users, secret.AuthorizedUserIDs)
				assert.ElementsMatch(t, roles, secret.AuthorizedRoleIDs)
				assert.Equal(t, userID1, secret.OwnerUserId)
			}
		})
	}
}

func TestSecretsService_GetSecretValue(t *testing.T) {
	service := setupSecretsTestService(t)
	ctx := context.Background()

	tests := []struct {
		name               string
		path               string
		value              string
		accessAsUserNumber int
		createSecret       bool
		expectedError      bool
		expectedErrorText  string
	}{
		{
			name:               "access as owner",
			path:               "/test/secret1",
			value:              "test_value_1",
			accessAsUserNumber: 0,
			createSecret:       true,
		},
		{
			name:               "access as authorized user",
			path:               "/test/secret1",
			value:              "test_value_1",
			accessAsUserNumber: 1,
			createSecret:       true,
		},
		{
			name:               "access as authorized role",
			path:               "/test/secret1",
			value:              "test_value_1",
			accessAsUserNumber: 2,
			createSecret:       true,
		},
		{
			name:               "access as not authorized",
			path:               "/test/secret1",
			value:              "test_value_1",
			accessAsUserNumber: 3,
			createSecret:       true,
			expectedError:      true,
			expectedErrorText:  "you do not have access to this secret",
		},
		{
			name:               "non-existing secret",
			path:               "/test/secret1",
			value:              "test_value_1",
			accessAsUserNumber: 0,
			createSecret:       false,
			expectedError:      true,
			expectedErrorText:  "you do not have access to this secret",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearSecretsServiceTables(t, ctx)

			userID1, err := service.usersRepository.CreateUser(ctx, "user1", "password1")
			require.NoError(t, err, "failed to create user1")
			userID2, err := service.usersRepository.CreateUser(ctx, "user2", "password2")
			require.NoError(t, err, "failed to create user2")
			userID3, err := service.usersRepository.CreateUser(ctx, "user3", "password3")
			require.NoError(t, err, "failed to create user3")
			userID4, err := service.usersRepository.CreateUser(ctx, "user4", "password4")
			require.NoError(t, err, "failed to create user4")
			roleID1, err := service.rolesRepository.CreateRole(ctx, "role1", "#FFFFFF")
			require.NoError(t, err, "failed to create role1")
			err = service.userRolesRepository.AssignRoleToUser(ctx, userID3, roleID1)
			require.NoError(t, err, "failed to assign role1 to user1")

			users := []int{userID1, userID2, userID3, userID4}

			if tt.createSecret {
				_, err = service.secretsRepository.CreateSecret(ctx, service.keyManager, tt.path, userID1, tt.value)
				require.NoError(t, err, "failed to create secret")

				err = service.secretsAccessRepository.GrantUserAccess(ctx, tt.path, userID2)
				require.NoError(t, err, "failed to grant users access")
				err = service.secretsAccessRepository.GrantRoleAccess(ctx, tt.path, roleID1)
				require.NoError(t, err, "failed to grant roles access")

			}
			ctx = context.WithValue(ctx, "token", &JwtClaims{
				UserID: users[tt.accessAsUserNumber],
			})

			payload := &gensecrets.GetSecretValuePayload{
				Path: base64.StdEncoding.EncodeToString([]byte(tt.path)),
			}

			getSecretValueResultPayload, err := service.GetSecretValue(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, getSecretValueResultPayload)
				assert.Equal(t, tt.path, *getSecretValueResultPayload.Path)
				assert.Equal(t, tt.value, *getSecretValueResultPayload.Value)
			}
		})
	}
}
