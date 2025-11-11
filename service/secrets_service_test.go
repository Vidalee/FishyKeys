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
		name                   string
		path                   string
		initialValue           string
		newValue               string
		initialAuthorizedUsers []int
		initialAuthorizedRoles []int
		newAuthorizedUsers     []int
		newAuthorizedRoles     []int
		updateAsUserNumber     int
		isAdmin                bool
		createSecret           bool
		expectedError          bool
		expectedErrorText      string
		verifySecretValue      bool
		verifyAuthorizedUsers  []int
		verifyAuthorizedRoles  []int
	}{
		{
			name:                   "owner who is also admin updates secret value successfully",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1}, // admin role
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     0,
			isAdmin:                true, // owner is also admin
			createSecret:           true,
			verifySecretValue:      true,
			verifyAuthorizedUsers:  []int{},
			verifyAuthorizedRoles:  []int{1},
		},
		{
			name:                   "owner without admin role cannot update",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     0,
			isAdmin:                false, // owner is not admin
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "only the owner or an admin can update the secret",
		},
		{
			name:                   "owner who is admin adds new users and roles",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{1, 2}, // userID2, userID3
			newAuthorizedRoles:     []int{1, 2}, // admin role + roleID1
			updateAsUserNumber:     0,
			isAdmin:                true, // owner is also admin
			createSecret:           true,
			verifySecretValue:      true,
			verifyAuthorizedUsers:  []int{1, 2},
			verifyAuthorizedRoles:  []int{1, 2},
		},
		{
			name:                   "owner who is admin removes users and roles",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{1, 2}, // userID2, userID3
			initialAuthorizedRoles: []int{1, 2}, // admin role + roleID1
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1}, // only admin role
			updateAsUserNumber:     0,
			isAdmin:                true, // owner is also admin
			createSecret:           true,
			verifySecretValue:      true,
			verifyAuthorizedUsers:  []int{},
			verifyAuthorizedRoles:  []int{1},
		},
		{
			name:                   "owner who is admin removes access from everyone except admin",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{1, 2, 3}, // userID2, userID3, userID4
			initialAuthorizedRoles: []int{1, 2},    // admin role + roleID1
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1}, // only admin role
			updateAsUserNumber:     0,
			isAdmin:                true, // owner is also admin
			createSecret:           true,
			verifySecretValue:      true,
			verifyAuthorizedUsers:  []int{},
			verifyAuthorizedRoles:  []int{1},
		},
		{
			name:                   "admin without being owner cannot update someone else's secret",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{1},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     4, // userID5 (admin, but not owner)
			isAdmin:                true,
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "only the owner or an admin can update the secret",
		},
		{
			name:                   "user with access but not owner or admin cannot update",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{2}, // userID2 has access
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{1},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     1, // userID2 (has access but not owner/admin)
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "only the owner or an admin can update the secret",
		},
		{
			name:                   "user without access cannot update",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     3, // userID4 (no access)
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "you do not have access to this secret",
		},
		{
			name:                   "owner who is admin cannot remove admin role",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1, 2}, // admin role + roleID1
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{2}, // trying to remove admin role
			updateAsUserNumber:     0,
			isAdmin:                true, // owner is also admin
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "admin role (ID 1) must always have access to the secret",
		},
		{
			name:                   "admin without being owner cannot update (fails before admin role check)",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1, 2}, // admin role + roleID1
			newAuthorizedRoles:     []int{2},    // trying to remove admin role
			updateAsUserNumber:     4,           // userID5 (admin, but not owner)
			isAdmin:                true,
			createSecret:           true,
			expectedError:          true,
			expectedErrorText:      "only the owner or an admin can update the secret",
		},
		{
			name:                   "non-existent secret",
			path:                   "/test/nonexistent",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     0,
			createSecret:           false,
			expectedError:          true,
			expectedErrorText:      "you do not have access to this secret",
		},
		{
			name:                   "invalid path encoding",
			path:                   "invalid_base64!@#",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{},
			initialAuthorizedRoles: []int{1},
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     0,
			createSecret:           false,
			expectedError:          true,
			expectedErrorText:      "invalid path encoding",
		},
		{
			name:                   "owner who is admin updates with empty authorized users and only admin role",
			path:                   "/test/secret1",
			initialValue:           "initial_value",
			newValue:               "updated_value",
			initialAuthorizedUsers: []int{1, 2},
			initialAuthorizedRoles: []int{1, 2},
			newAuthorizedUsers:     []int{},
			newAuthorizedRoles:     []int{1},
			updateAsUserNumber:     0,
			isAdmin:                true, // owner is also admin
			createSecret:           true,
			verifySecretValue:      true,
			verifyAuthorizedUsers:  []int{},
			verifyAuthorizedRoles:  []int{1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearSecretsServiceTables(t, ctx)

			// Create users
			userID1, err := service.usersRepository.CreateUser(ctx, "user1", "password1")
			require.NoError(t, err, "failed to create user1")
			userID2, err := service.usersRepository.CreateUser(ctx, "user2", "password2")
			require.NoError(t, err, "failed to create user2")
			userID3, err := service.usersRepository.CreateUser(ctx, "user3", "password3")
			require.NoError(t, err, "failed to create user3")
			userID4, err := service.usersRepository.CreateUser(ctx, "user4", "password4")
			require.NoError(t, err, "failed to create user4")
			userID5, err := service.usersRepository.CreateUser(ctx, "user5", "password5")
			require.NoError(t, err, "failed to create user5")

			// Get admin role (created by migration, preserved by ClearTable)
			adminRole, err := service.rolesRepository.GetRoleByName(ctx, "admin")
			require.NoError(t, err, "failed to get admin role")
			adminRoleID := adminRole.ID

			roleID1, err := service.rolesRepository.CreateRole(ctx, "role1", "#FFFFFF")
			require.NoError(t, err, "failed to create role1")

			// Assign admin role to users if needed
			if tt.isAdmin {
				// If updateAsUserNumber is 0, assign admin to user1 (owner)
				// If updateAsUserNumber is 4, assign admin to user5
				if tt.updateAsUserNumber == 0 {
					err = service.userRolesRepository.AssignRoleToUser(ctx, userID1, adminRoleID)
					require.NoError(t, err, "failed to assign admin role to user1")
				} else if tt.updateAsUserNumber == 4 {
					err = service.userRolesRepository.AssignRoleToUser(ctx, userID5, adminRoleID)
					require.NoError(t, err, "failed to assign admin role to user5")
				}
			}

			users := []int{userID1, userID2, userID3, userID4, userID5}

			// Map test indices to actual user IDs
			// initialAuthorizedUsers uses indices 1, 2, 3 which map to userID2, userID3, userID4
			// newAuthorizedUsers uses indices 1, 2 which map to userID2, userID3
			mapUserIndices := func(indices []int) []int {
				mapped := make([]int, 0, len(indices))
				for _, idx := range indices {
					if idx > 0 && idx <= len(users) {
						mapped = append(mapped, users[idx-1])
					}
				}
				return mapped
			}

			// Map role indices: 1 = adminRoleID, 2 = roleID1
			mapRoleIndices := func(indices []int) []int {
				mapped := make([]int, 0, len(indices))
				for _, idx := range indices {
					if idx == 1 {
						mapped = append(mapped, adminRoleID)
					} else if idx == 2 {
						mapped = append(mapped, roleID1)
					}
				}
				return mapped
			}

			initialUsers := mapUserIndices(tt.initialAuthorizedUsers)
			initialRoles := mapRoleIndices(tt.initialAuthorizedRoles)
			newUsers := mapUserIndices(tt.newAuthorizedUsers)
			newRoles := mapRoleIndices(tt.newAuthorizedRoles)
			verifyUsers := mapUserIndices(tt.verifyAuthorizedUsers)
			verifyRoles := mapRoleIndices(tt.verifyAuthorizedRoles)

			if tt.createSecret {
				_, err = service.secretsRepository.CreateSecret(ctx, service.keyManager, tt.path, userID1, tt.initialValue)
				require.NoError(t, err, "failed to create secret")

				// Grant initial access
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
				UserID: users[tt.updateAsUserNumber],
			})

			payload := &gensecrets.UpdateSecretPayload{
				Path:            base64.StdEncoding.EncodeToString([]byte(tt.path)),
				Value:           tt.newValue,
				AuthorizedUsers: newUsers,
				AuthorizedRoles: newRoles,
			}

			// Handle invalid path encoding case
			if tt.name == "invalid path encoding" {
				payload.Path = tt.path
			}

			err = service.UpdateSecret(ctxWithToken, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorText)
			} else {
				assert.NoError(t, err)

				if tt.verifySecretValue {
					secret, err := service.secretsRepository.GetSecretByPath(ctx, service.keyManager, tt.path)
					assert.NoError(t, err)
					assert.NotNil(t, secret)
					assert.Equal(t, tt.newValue, secret.DecryptedValue, "secret value should be updated")

					authorizedUsers, authorizedRoles, err := service.secretsAccessRepository.GetAccessesBySecretPath(ctx, tt.path)
					assert.NoError(t, err)
					assert.ElementsMatch(t, verifyUsers, authorizedUsers, "authorized users should match")
					assert.ElementsMatch(t, verifyRoles, authorizedRoles, "authorized roles should match")
				}
			}
		})
	}
}
