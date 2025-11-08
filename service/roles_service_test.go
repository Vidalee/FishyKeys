package service

import (
	"context"
	"testing"

	genroles "github.com/Vidalee/FishyKeys/gen/roles"
	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupRolesTestService(t *testing.T) *RolesService {
	rolesRepo := repository.NewRolesRepository(testDB)
	usersRepo := repository.NewUsersRepository(testDB)
	userRolesRepo := repository.NewUserRolesRepository(testDB)
	return NewRolesService(rolesRepo, usersRepo, userRolesRepo)
}

func clearRolesServiceTables(t *testing.T, ctx context.Context) {
	err := testutil.ClearTable(ctx, "user_roles")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "users")
	require.NoError(t, err)
	err = testutil.ClearTable(ctx, "roles")
	require.NoError(t, err)
}

func TestUsersService_ListRoles(t *testing.T) {
	service := setupRolesTestService(t)
	ctx := context.Background()
	clearRolesServiceTables(t, ctx)

	roles, err := service.ListRoles(ctx)
	assert.NoError(t, err)
	// 1 since there is a system role
	assert.Len(t, roles, 1, "expected no roles after clearing table")

	testRoles := []struct {
		Name  string
		Color string
	}{
		{"role1", "#0000FF"},
		{"role2", "#FF0000"},
	}

	var createdRoleIDs []int

	for _, r := range testRoles {
		roleID, err := service.rolesRepository.CreateRole(ctx, r.Name, r.Color)
		require.NoError(t, err, "failed to create role: %s", r.Name)
		createdRoleIDs = append(createdRoleIDs, roleID)
	}

	roles, err = service.ListRoles(ctx)
	assert.NoError(t, err, "failed to list roles")
	expectedCount := 1 + len(testRoles) // 1 for the system role
	assert.Len(t, roles, expectedCount, "unexpected number of roles")

	for i, r := range testRoles {
		role := roles[i+1]
		assert.Equal(t, createdRoleIDs[i], role.ID, "ID mismatch for role %s", r.Name)
		assert.Equal(t, r.Name, role.Name, "name mismatch for role %s", r.Name)
		assert.Equal(t, r.Color, role.Color, "color mismatch for role %s", r.Name)
	}
}

func TestRolesService_CreateRole(t *testing.T) {
	service := setupRolesTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		roleName          string
		roleColor         string
		expectedError     bool
		expectedErrorText string
		doDuplicate       bool
	}{
		{
			name:          "valid role",
			roleName:      "user",
			roleColor:     "#FF0000",
			expectedError: false,
		},
		{
			name:              "duplicate role name",
			roleName:          "admin",
			roleColor:         "#00FF00",
			expectedError:     true,
			expectedErrorText: "role with name admin already exists",
			doDuplicate:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearRolesServiceTables(t, ctx)

			payload := &genroles.CreateRolePayload{
				Name:  tt.roleName,
				Color: tt.roleColor,
			}

			result, err := service.CreateRole(ctx, payload)
			if tt.doDuplicate {
				result, err = service.CreateRole(ctx, payload)
			}

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedErrorText)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotNil(t, result.Name)
				assert.Equal(t, tt.roleName, result.Name)
				assert.Equal(t, tt.roleColor, result.Color)
				assert.NotEqual(t, 0, result.ID, "role ID should not be zero")

				role, err := service.rolesRepository.GetRoleByID(ctx, result.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.roleName, role.Name)
				assert.Equal(t, tt.roleColor, role.Color)
			}
		})
	}
}

func TestRolesService_DeleteRole(t *testing.T) {
	service := setupRolesTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		roleName          string
		roleColor         string
		createRole        bool
		roleID            int
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "nonexistent role",
			createRole:        false,
			roleID:            999,
			expectedError:     true,
			expectedErrorText: "role not found",
		},
		{
			name:          "existing role",
			createRole:    true,
			roleName:      "testrole",
			roleColor:     "#0000FF",
			expectedError: false,
		},
		{
			name:              "delete admin role",
			createRole:        false,
			roleID:            1,
			expectedError:     true,
			expectedErrorText: "cannot delete admin role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearRolesServiceTables(t, ctx)

			var roleID int
			if tt.createRole {
				createPayload := &genroles.CreateRolePayload{
					Name:  tt.roleName,
					Color: tt.roleColor,
				}
				result, err := service.CreateRole(ctx, createPayload)
				require.NoError(t, err, "failed to create role for delete test")
				roleID = result.ID
			} else {
				roleID = tt.roleID
			}

			payload := &genroles.DeleteRolePayload{
				ID: roleID,
			}

			err := service.DeleteRole(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorText)
			} else {
				assert.NoError(t, err)

				_, err := service.rolesRepository.GetRoleByID(ctx, roleID)
				assert.Error(t, err)
				assert.Equal(t, repository.ErrRoleNotFound, err, "expected role to be deleted")
			}
		})
	}
}

func TestRolesService_AssignRoleToUser(t *testing.T) {
	service := setupRolesTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		createUser        bool
		createRole        bool
		userID            int
		roleID            int
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "nonexistent user",
			createUser:        false,
			createRole:        true,
			userID:            999,
			expectedError:     true,
			expectedErrorText: "user not found",
		},
		{
			name:              "nonexistent role",
			createUser:        true,
			createRole:        false,
			roleID:            999,
			expectedError:     true,
			expectedErrorText: "role not found",
		},
		{
			name:          "valid assignment",
			createUser:    true,
			createRole:    true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearRolesServiceTables(t, ctx)

			var userID int
			var roleID int

			if tt.createUser {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				require.NoError(t, err, "failed to hash password")
				userID, err = service.userRepository.CreateUser(ctx, "testuser", string(hashedPassword))
				require.NoError(t, err, "failed to create user")
			} else {
				userID = tt.userID
			}

			if tt.createRole {
				createdRoleID, err := service.rolesRepository.CreateRole(ctx, "testrole", "#FF0000")
				require.NoError(t, err, "failed to create role")
				roleID = createdRoleID
			} else {
				roleID = tt.roleID
			}

			payload := &genroles.AssignRoleToUserPayload{
				UserID: userID,
				RoleID: roleID,
			}

			err := service.AssignRoleToUser(ctx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorText)
			} else {
				assert.NoError(t, err)

				// Verify the role was assigned
				roleIDs, err := service.userRolesRepository.GetUserRoleIDs(ctx, userID)
				assert.NoError(t, err)
				assert.Contains(t, roleIDs, roleID, "role should be assigned to user")
			}
		})
	}
}

func TestRolesService_UnassignRoleToUser(t *testing.T) {
	service := setupRolesTestService(t)
	ctx := context.Background()

	tests := []struct {
		name              string
		createUser        bool
		createRole        bool
		assignRole        bool
		userID            int
		roleID            int
		setToken          bool
		tokenUserID       int
		isAdminRole       bool
		expectedError     bool
		expectedErrorText string
	}{
		{
			name:              "nonexistent user",
			createUser:        false,
			createRole:        true,
			assignRole:        false,
			userID:            999,
			expectedError:     true,
			expectedErrorText: "user not found",
		},
		{
			name:              "nonexistent role",
			createUser:        true,
			createRole:        false,
			assignRole:        false,
			roleID:            999,
			expectedError:     true,
			expectedErrorText: "role not found",
		},
		{
			name:          "valid unassignment",
			createUser:    true,
			createRole:    true,
			assignRole:    true,
			expectedError: false,
		},
		{
			name:              "unassign admin role from yourself",
			createUser:        true,
			createRole:        false,
			assignRole:        true,
			roleID:            1, // admin role
			setToken:          true,
			isAdminRole:       true,
			expectedError:     true,
			expectedErrorText: "cannot unassign admin role from yourself",
		},
		{
			name:          "unassign admin role from different user",
			createUser:    true,
			createRole:    false,
			assignRole:    true,
			roleID:        1, // admin role
			setToken:      true,
			tokenUserID:   999, // different user
			isAdminRole:   true,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearRolesServiceTables(t, ctx)

			var userID int
			var roleID int

			if tt.createUser {
				hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
				require.NoError(t, err, "failed to hash password")
				userID, err = service.userRepository.CreateUser(ctx, "testuser", string(hashedPassword))
				require.NoError(t, err, "failed to create user")
			} else {
				userID = tt.userID
			}

			if tt.createRole {
				createdRoleID, err := service.rolesRepository.CreateRole(ctx, "testrole", "#FF0000")
				require.NoError(t, err, "failed to create role")
				roleID = createdRoleID
			} else {
				roleID = tt.roleID
			}

			if tt.assignRole {
				err := service.userRolesRepository.AssignRoleToUser(ctx, userID, roleID)
				require.NoError(t, err, "failed to assign role to user")
			}

			testCtx := ctx
			if tt.setToken {
				tokenUserID := userID
				if tt.tokenUserID != 0 {
					tokenUserID = tt.tokenUserID
				}
				testCtx = context.WithValue(ctx, "token", &JwtClaims{
					UserID: tokenUserID,
				})
			}

			payload := &genroles.UnassignRoleToUserPayload{
				UserID: userID,
				RoleID: roleID,
			}

			err := service.UnassignRoleToUser(testCtx, payload)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrorText)
			} else {
				assert.NoError(t, err)

				// Verify the role was unassigned
				roleIDs, err := service.userRolesRepository.GetUserRoleIDs(ctx, userID)
				assert.NoError(t, err)
				assert.NotContains(t, roleIDs, roleID, "role should be unassigned from user")
			}
		})
	}
}
