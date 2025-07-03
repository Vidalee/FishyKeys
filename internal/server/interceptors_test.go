package server

import (
	"context"
	"errors"
	"testing"

	goa "goa.design/goa/v3/pkg"

	"github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	repositorymocks "github.com/Vidalee/FishyKeys/gen/mocks"
)

func TestServerInterceptors_Authentified(t *testing.T) {
	interceptor := &ServerUsersInterceptors{}

	t.Run("should return unauthorized when token is missing", func(t *testing.T) {
		ctx := context.Background()

		nextCalled := false
		next := func(ctx context.Context, req any) (any, error) {
			nextCalled = true
			return nil, nil
		}

		info := &users.AuthentifiedInfo{}

		resp, err := interceptor.Authentified(ctx, info, next)

		require.Error(t, err)
		assert.Nil(t, resp)

		var goaErr *goa.ServiceError
		ok := errors.As(err, &goaErr)
		assert.True(t, ok, "error should be a ServiceError")

		assert.Equal(t, "unauthorized", goaErr.Name)
		assert.Equal(t, "you need to be authenticated to access this endpoint", goaErr.Message)

		assert.False(t, nextCalled, "next endpoint should not have been called")
	})

	t.Run("should call next when token is present", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), "token", "valid-token")

		nextCalled := false
		next := func(ctx context.Context, req any) (any, error) {
			nextCalled = true
			return "success", nil
		}

		info := &users.AuthentifiedInfo{}

		resp, err := interceptor.Authentified(ctx, info, next)

		require.NoError(t, err)
		assert.Equal(t, "success", resp)
		assert.True(t, nextCalled, "next endpoint should have been called")
	})
}

func TestServerInterceptors_IsAdmin(t *testing.T) {
	jwtClaims := &service.JwtClaims{UserID: 42}
	ctxWithToken := context.WithValue(context.Background(), "token", jwtClaims)

	t.Run("should return unauthorized when token is missing", func(t *testing.T) {
		interceptor := &ServerUsersInterceptors{}

		next := func(ctx context.Context, req any) (any, error) {
			return "should not call", nil
		}

		info := &users.IsAdminInfo{}

		resp, err := interceptor.IsAdmin(context.Background(), info, next)

		require.Error(t, err)
		assert.Nil(t, resp)

		var goaErr *goa.ServiceError
		ok := errors.As(err, &goaErr)
		assert.True(t, ok, "error should be a ServiceError")

		assert.Equal(t, "unauthorized", goaErr.Name)
	})

	t.Run("should return internal error if userRolesRepository fails", func(t *testing.T) {
		mockUserRoles := repositorymocks.NewMockUserRolesRepository(t)
		mockUserRoles.On("GetUserRoleIDs", mock.Anything, 42).Return([]int{}, errors.New("db error"))

		interceptor := &ServerUsersInterceptors{userRolesRepository: mockUserRoles}

		next := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		info := &users.IsAdminInfo{}

		resp, err := interceptor.IsAdmin(ctxWithToken, info, next)

		require.Error(t, err)
		assert.Nil(t, resp)

		var goaErr *goa.ServiceError
		ok := errors.As(err, &goaErr)
		assert.True(t, ok, "error should be a ServiceError")

		assert.Equal(t, "internal_error", goaErr.Name)
	})

	t.Run("should return internal error if rolesRepository fails", func(t *testing.T) {
		mockUserRoles := repositorymocks.NewMockUserRolesRepository(t)
		mockRoles := repositorymocks.NewMockRolesRepository(t)

		mockUserRoles.On("GetUserRoleIDs", mock.Anything, 42).Return([]int{1, 2}, nil)
		mockRoles.On("GetRolesByIDs", mock.Anything, []int{1, 2}).Return([]repository.Role{}, errors.New("db error"))

		interceptor := &ServerUsersInterceptors{
			userRolesRepository: mockUserRoles,
			rolesRepository:     mockRoles,
		}

		next := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		info := &users.IsAdminInfo{}

		resp, err := interceptor.IsAdmin(ctxWithToken, info, next)

		require.Error(t, err)
		assert.Nil(t, resp)

		var goaErr *goa.ServiceError
		ok := errors.As(err, &goaErr)
		assert.True(t, ok, "error should be a ServiceError")

		assert.Equal(t, "internal_error", goaErr.Name)
	})

	t.Run("should return forbidden if user has no roles", func(t *testing.T) {
		mockUserRoles := repositorymocks.NewMockUserRolesRepository(t)
		mockRoles := repositorymocks.NewMockRolesRepository(t)

		mockUserRoles.On("GetUserRoleIDs", mock.Anything, 42).Return([]int{}, nil)
		mockRoles.On("GetRolesByIDs", mock.Anything, []int{}).Return([]repository.Role{}, nil)

		interceptor := &ServerUsersInterceptors{
			userRolesRepository: mockUserRoles,
			rolesRepository:     mockRoles,
		}

		next := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		info := &users.IsAdminInfo{}

		resp, err := interceptor.IsAdmin(ctxWithToken, info, next)

		require.Error(t, err)
		assert.Nil(t, resp)

		var goaErr *goa.ServiceError
		ok := errors.As(err, &goaErr)
		assert.True(t, ok, "error should be a ServiceError")

		assert.Equal(t, "forbidden", goaErr.Name)
	})

	t.Run("should return forbidden if user is not admin", func(t *testing.T) {
		mockUserRoles := repositorymocks.NewMockUserRolesRepository(t)
		mockRoles := repositorymocks.NewMockRolesRepository(t)

		mockUserRoles.On("GetUserRoleIDs", mock.Anything, 42).Return([]int{1}, nil)
		mockRoles.On("GetRolesByIDs", mock.Anything, []int{1}).Return([]repository.Role{{ID: 1, Admin: false}}, nil)

		interceptor := &ServerUsersInterceptors{
			userRolesRepository: mockUserRoles,
			rolesRepository:     mockRoles,
		}

		next := func(ctx context.Context, req any) (any, error) {
			return nil, nil
		}

		info := &users.IsAdminInfo{}

		resp, err := interceptor.IsAdmin(ctxWithToken, info, next)

		require.Error(t, err)
		assert.Nil(t, resp)

		var goaErr *goa.ServiceError
		ok := errors.As(err, &goaErr)
		assert.True(t, ok, "error should be a ServiceError")

		assert.Equal(t, "forbidden", goaErr.Name)
	})

	t.Run("should call next if user is admin", func(t *testing.T) {
		mockUserRoles := repositorymocks.NewMockUserRolesRepository(t)
		mockRoles := repositorymocks.NewMockRolesRepository(t)

		mockUserRoles.On("GetUserRoleIDs", mock.Anything, 42).Return([]int{1}, nil)
		mockRoles.On("GetRolesByIDs", mock.Anything, []int{1}).Return([]repository.Role{{ID: 1, Admin: true}}, nil)

		interceptor := &ServerUsersInterceptors{
			userRolesRepository: mockUserRoles,
			rolesRepository:     mockRoles,
		}

		nextCalled := false
		next := func(ctx context.Context, req any) (any, error) {
			nextCalled = true
			return "ok", nil
		}

		info := &users.IsAdminInfo{}

		resp, err := interceptor.IsAdmin(ctxWithToken, info, next)

		require.NoError(t, err)
		assert.Equal(t, "ok", resp)
		assert.True(t, nextCalled, "next endpoint should have been called")
	})
}
