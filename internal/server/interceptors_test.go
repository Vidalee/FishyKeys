package server

import (
	"context"
	"errors"
	goa "goa.design/goa/v3/pkg"
	"testing"

	"github.com/Vidalee/FishyKeys/gen/users"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerInterceptors_Authentified(t *testing.T) {
	interceptor := &ServerInterceptors{}

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
