package service

import (
	"context"
	"encoding/base64"
	"testing"

	gensecrets "github.com/Vidalee/FishyKeys/gen/secrets"

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
			err = service.userRolesRepository.AssignRoleToUser(ctx, userID2, roleID1)
			require.NoError(t, err, "failed to assign role1 to user2")

			var roles []int
			var users []int
			if tt.addCreatedUsersAndRoles {
				roles = []int{roleID1}
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

			payload := &gensecrets.OperatorGetSecretValuePayload{
				Path: base64.StdEncoding.EncodeToString([]byte(tt.path)),
			}

			getSecretValueResultPayload, err := service.OperatorGetSecretValue(ctx, payload)

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

func TestSecretsService_GetSecret(t *testing.T) {
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

			payload := &gensecrets.GetSecretPayload{
				Path: base64.StdEncoding.EncodeToString([]byte(tt.path)),
			}

			secretInfo, err := service.GetSecret(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, secretInfo)
				assert.Equal(t, tt.path, secretInfo.Path)
				assert.Equal(t, userID1, secretInfo.Owner.ID)
				assert.Equal(t, "user1", secretInfo.Owner.Username)
				assert.Equal(t, 1, len(secretInfo.AuthorizedUsers), "should have 1 authorized user")
				assert.Equal(t, userID2, secretInfo.AuthorizedUsers[0].ID)
				assert.Equal(t, "user2", secretInfo.AuthorizedUsers[0].Username)
				assert.Equal(t, 1, len(secretInfo.AuthorizedRoles), "should have 1 authorized role")
				assert.Equal(t, roleID1, secretInfo.AuthorizedRoles[0].ID)
				assert.Equal(t, "role1", secretInfo.AuthorizedRoles[0].Name)
				assert.NotEmpty(t, secretInfo.CreatedAt, "created at should not be empty")
				assert.NotEmpty(t, secretInfo.UpdatedAt, "updated at should not be empty")
			}
		})
	}
}

func TestSecretsService_ListSecrets(t *testing.T) {
	service := setupSecretsTestService(t)
	ctx := context.Background()
	clearSecretsServiceTables(t, ctx)

	userID1, err := service.usersRepository.CreateUser(ctx, "user1", "password1")
	require.NoError(t, err, "failed to create user1")
	userID2, err := service.usersRepository.CreateUser(ctx, "user2", "password2")
	require.NoError(t, err, "failed to create user2")
	userID3, err := service.usersRepository.CreateUser(ctx, "user3", "password3")
	require.NoError(t, err, "failed to create user3")
	roleID1, err := service.rolesRepository.CreateRole(ctx, "role1", "#FFFFFF")
	require.NoError(t, err, "failed to create role1")
	err = service.userRolesRepository.AssignRoleToUser(ctx, userID3, roleID1)
	require.NoError(t, err, "failed to assign role1 to user3")

	_, err = service.secretsRepository.CreateSecret(ctx, service.keyManager, "/owned/secret", userID1, "owned_value")
	require.NoError(t, err, "failed to create owned secret")
	_, err = service.secretsRepository.CreateSecret(ctx, service.keyManager, "/user/secret", userID1, "user_value")
	require.NoError(t, err, "failed to create user secret")
	_, err = service.secretsRepository.CreateSecret(ctx, service.keyManager, "/role/secret", userID1, "role_value")
	require.NoError(t, err, "failed to create role secret")

	err = service.secretsAccessRepository.GrantUserAccess(ctx, "/user/secret", userID2)
	require.NoError(t, err, "failed to grant user2 access to /user/secret")
	err = service.secretsAccessRepository.GrantRoleAccess(ctx, "/role/secret", roleID1)
	require.NoError(t, err, "failed to grant role1 access to /role/secret")

	t.Run("owner sees all their secrets", func(t *testing.T) {
		ctxWithToken := context.WithValue(ctx, "token", &JwtClaims{UserID: userID1})
		secrets, err := service.ListSecrets(ctxWithToken)
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
		paths := make([]string, 0, len(secrets))
		for _, s := range secrets {
			paths = append(paths, s.Path)
			assert.Equal(t, s.Owner.ID, userID1)
		}
		assert.ElementsMatch(t, []string{"/owned/secret", "/user/secret", "/role/secret"}, paths)
	})

	t.Run("user with direct access sees only authorized secret", func(t *testing.T) {
		ctxWithToken := context.WithValue(ctx, "token", &JwtClaims{UserID: userID2})
		secrets, err := service.ListSecrets(ctxWithToken)
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
		paths := make([]string, 0, len(secrets))
		for _, s := range secrets {
			paths = append(paths, s.Path)
			assert.Equal(t, s.Owner.ID, userID1)
			assert.ElementsMatch(t, usersToIds(s.Users), []int{userID2})
			assert.ElementsMatch(t, rolesToIds(s.Roles), []string{})
		}
		assert.ElementsMatch(t, []string{"/user/secret"}, paths)
	})

	t.Run("user with role access sees only role secret", func(t *testing.T) {
		ctxWithToken := context.WithValue(ctx, "token", &JwtClaims{UserID: userID3})
		secrets, err := service.ListSecrets(ctxWithToken)
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
		paths := make([]string, 0, len(secrets))
		for _, s := range secrets {
			paths = append(paths, s.Path)
			assert.Equal(t, s.Owner.ID, userID1)
			assert.ElementsMatch(t, usersToIds(s.Users), []int{})
			assert.ElementsMatch(t, rolesToIds(s.Roles), []int{roleID1})
		}
		assert.ElementsMatch(t, []string{"/role/secret"}, paths)
	})

	t.Run("user with no access sees nothing", func(t *testing.T) {
		ctxWithToken := context.WithValue(ctx, "token", &JwtClaims{UserID: 9999}) // non-existent user
		secrets, err := service.ListSecrets(ctxWithToken)
		assert.NoError(t, err)
		assert.NotNil(t, secrets)
		assert.Empty(t, secrets)
	})
}

func usersToIds(users []*gensecrets.User) []int {
	ids := make([]int, 0, len(users))
	for _, u := range users {
		ids = append(ids, u.ID)
	}
	return ids
}

func rolesToIds(roles []*gensecrets.Role) []int {
	ids := make([]int, 0, len(roles))
	for _, r := range roles {
		ids = append(ids, r.ID)
	}
	return ids
}

func TestSecretsService_UpdateSecret(t *testing.T) {
	service := setupSecretsTestService(t)
	ctx := context.Background()

	tests := []struct {
		name                    string
		path                    string
		initialValue            string
		updatedValue            string
		accessAsUserNumber      int
		initialAuthorizedUsers  []int
		initialAuthorizedRoles  []int
		updateAuthorizedUsers   []int
		updateAuthorizedRoles   []int
		expectedAuthorizedUsers []int
		expectedAuthorizedRoles []int
		createSecret            bool
		expectedError           bool
		expectedErrorText       string
	}{
		{
			name:                    "owner updates value only",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      0,
			initialAuthorizedUsers:  []int{},
			initialAuthorizedRoles:  []int{},
			updateAuthorizedUsers:   []int{},
			updateAuthorizedRoles:   []int{},
			expectedAuthorizedUsers: []int{},
			expectedAuthorizedRoles: []int{},
			createSecret:            true,
		},
		{
			name:                    "owner adds user access",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      0,
			initialAuthorizedUsers:  []int{},
			initialAuthorizedRoles:  []int{},
			updateAuthorizedUsers:   []int{1},
			updateAuthorizedRoles:   []int{},
			expectedAuthorizedUsers: []int{1},
			expectedAuthorizedRoles: []int{},
			createSecret:            true,
		},
		{
			name:                    "owner removes user access",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      0,
			initialAuthorizedUsers:  []int{1},
			initialAuthorizedRoles:  []int{},
			updateAuthorizedUsers:   []int{1},
			updateAuthorizedRoles:   []int{},
			expectedAuthorizedUsers: []int{},
			expectedAuthorizedRoles: []int{},
			createSecret:            true,
		},
		{
			name:                    "owner adds role access",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      0,
			initialAuthorizedUsers:  []int{},
			initialAuthorizedRoles:  []int{},
			updateAuthorizedUsers:   []int{},
			updateAuthorizedRoles:   []int{0},
			expectedAuthorizedUsers: []int{},
			expectedAuthorizedRoles: []int{0},
			createSecret:            true,
		},
		{
			name:                    "owner removes role access",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      0,
			initialAuthorizedUsers:  []int{},
			initialAuthorizedRoles:  []int{0},
			updateAuthorizedUsers:   []int{},
			updateAuthorizedRoles:   []int{0},
			expectedAuthorizedUsers: []int{},
			expectedAuthorizedRoles: []int{},
			createSecret:            true,
		},
		{
			name:                    "owner toggles multiple users and roles",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      0,
			initialAuthorizedUsers:  []int{1},
			initialAuthorizedRoles:  []int{0},
			updateAuthorizedUsers:   []int{1, 2}, // remove 1, add 2
			updateAuthorizedRoles:   []int{0, 1}, // remove 0, add 1
			expectedAuthorizedUsers: []int{2},
			expectedAuthorizedRoles: []int{1},
			createSecret:            true,
		},
		{
			name:                    "authorized user updates value",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      1,
			initialAuthorizedUsers:  []int{1},
			initialAuthorizedRoles:  []int{},
			updateAuthorizedUsers:   []int{},
			updateAuthorizedRoles:   []int{},
			expectedAuthorizedUsers: []int{1},
			expectedAuthorizedRoles: []int{},
			createSecret:            true,
		},
		{
			name:                   "authorized user cannot update secret they don't have access to",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			updatedValue:           "updated_value",
			accessAsUserNumber:     2,
			initialAuthorizedUsers: []int{1},
			initialAuthorizedRoles: []int{},
			updateAuthorizedUsers:  []int{},
			updateAuthorizedRoles:  []int{},
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "you do not have access to this secret",
		},
		{
			name:                    "user with role access can update",
			path:                    "/test/secret1",
			initialValue:            "initial_value",
			updatedValue:            "updated_value",
			accessAsUserNumber:      2,
			initialAuthorizedUsers:  []int{},
			initialAuthorizedRoles:  []int{0},
			updateAuthorizedUsers:   []int{},
			updateAuthorizedRoles:   []int{},
			expectedAuthorizedUsers: []int{},
			expectedAuthorizedRoles: []int{0},
			createSecret:            true,
		},
		{
			name:                   "non-existent secret",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			updatedValue:           "updated_value",
			accessAsUserNumber:     0,
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{},
			updateAuthorizedUsers:  []int{},
			updateAuthorizedRoles:  []int{},
			createSecret:           false,
			expectedError:          true,
			expectedErrorText:      "you do not have access to this secret",
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
			roleID1, err := service.rolesRepository.CreateRole(ctx, "role1", "#FFFFFF")
			require.NoError(t, err, "failed to create role1")
			roleID2, err := service.rolesRepository.CreateRole(ctx, "role2", "#000000")
			require.NoError(t, err, "failed to create role2")
			err = service.userRolesRepository.AssignRoleToUser(ctx, userID3, roleID1)
			require.NoError(t, err, "failed to assign role1 to user3")

			users := []int{userID1, userID2, userID3}
			roles := []int{roleID1, roleID2}

			if tt.createSecret {
				_, err = service.secretsRepository.CreateSecret(ctx, service.keyManager, tt.path, userID1, tt.initialValue)
				require.NoError(t, err, "failed to create secret")

				// Map test indices to actual IDs
				initialUsers := make([]int, 0)
				for _, idx := range tt.initialAuthorizedUsers {
					if idx < len(users) {
						initialUsers = append(initialUsers, users[idx])
					}
				}
				initialRoles := make([]int, 0)
				for _, idx := range tt.initialAuthorizedRoles {
					if idx < len(roles) {
						initialRoles = append(initialRoles, roles[idx])
					}
				}

				if len(initialUsers) > 0 {
					err = service.secretsAccessRepository.GrantUsersAccess(ctx, tt.path, initialUsers)
					require.NoError(t, err, "failed to grant initial user access")
				}
				if len(initialRoles) > 0 {
					err = service.secretsAccessRepository.GrantRolesAccess(ctx, tt.path, initialRoles)
					require.NoError(t, err, "failed to grant initial role access")
				}
			}

			ctxWithToken := context.WithValue(ctx, "token", &JwtClaims{
				UserID: users[tt.accessAsUserNumber],
			})

			// Map test indices to actual IDs for update payload
			updateUsers := make([]int, 0)
			for _, idx := range tt.updateAuthorizedUsers {
				if idx < len(users) {
					updateUsers = append(updateUsers, users[idx])
				}
			}
			updateRoles := make([]int, 0)
			for _, idx := range tt.updateAuthorizedRoles {
				if idx < len(roles) {
					updateRoles = append(updateRoles, roles[idx])
				}
			}

			payload := &gensecrets.UpdateSecretPayload{
				Path:            tt.path,
				Value:           tt.updatedValue,
				AuthorizedUsers: updateUsers,
				AuthorizedRoles: updateRoles,
			}

			err = service.UpdateSecret(ctxWithToken, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorText, err.Error())
			} else {
				assert.NoError(t, err)

				// Verify secret value was updated
				secret, err := service.secretsRepository.GetSecretByPath(ctx, service.keyManager, tt.path)
				require.NoError(t, err)
				assert.Equal(t, tt.updatedValue, secret.DecryptedValue)

				// Verify access was updated correctly
				authorizedUserIds, authorizedRoleIds, err := service.secretsAccessRepository.GetAccessesBySecretPath(ctx, tt.path)
				require.NoError(t, err)

				// Map expected indices to actual IDs
				expectedUsers := make([]int, 0)
				for _, idx := range tt.expectedAuthorizedUsers {
					if idx < len(users) {
						expectedUsers = append(expectedUsers, users[idx])
					}
				}
				expectedRoles := make([]int, 0)
				for _, idx := range tt.expectedAuthorizedRoles {
					if idx < len(roles) {
						expectedRoles = append(expectedRoles, roles[idx])
					}
				}

				assert.ElementsMatch(t, expectedUsers, authorizedUserIds, "authorized users should match")
				assert.ElementsMatch(t, expectedRoles, authorizedRoleIds, "authorized roles should match")
			}
		})
	}
}
