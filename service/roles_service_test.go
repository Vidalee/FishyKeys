package service

import (
	"context"
	"testing"

	"github.com/Vidalee/FishyKeys/internal/testutil"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRolesTestService(t *testing.T) *RolesService {
	rolesRepo := repository.NewRolesRepository(testDB)
	return NewRolesService(rolesRepo)
}

func clearRolesServiceTables(t *testing.T, ctx context.Context) {
	err := testutil.ClearTable(ctx, "roles")
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

	// Define roles to test
	testRoles := []struct {
		Name  string
		Color string
	}{
		{"role1", "#0000FF"},
		{"role2", "#FF0000"},
	}

	var createdRoleIDs []int

	// Create roles
	for _, r := range testRoles {
		roleID, err := service.rolesRepository.CreateRole(ctx, r.Name, r.Color)
		require.NoError(t, err, "failed to create role: %s", r.Name)
		createdRoleIDs = append(createdRoleIDs, roleID)
	}

	// Fetch all roles and verify
	roles, err = service.ListRoles(ctx)
	assert.NoError(t, err, "failed to list roles")
	expectedCount := 1 + len(testRoles) // 1 for the system role
	assert.Len(t, roles, expectedCount, "unexpected number of roles")

	// Verify each created role (skip index 0 assuming it's the system role)
	for i, r := range testRoles {
		role := roles[i+1]
		assert.Equal(t, createdRoleIDs[i], role.ID, "ID mismatch for role %s", r.Name)
		assert.Equal(t, r.Name, role.Name, "name mismatch for role %s", r.Name)
		assert.Equal(t, r.Color, role.Color, "color mismatch for role %s", r.Name)
	}
}
