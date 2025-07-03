package server

import (
	"context"
	"fmt"
	"github.com/Vidalee/FishyKeys/gen/secrets"
	"github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/repository"
	"github.com/Vidalee/FishyKeys/service"
	goa "goa.design/goa/v3/pkg"
)

type ServerUsersInterceptors struct {
	rolesRepository     repository.RolesRepository
	userRolesRepository repository.UserRolesRepository
}

func (i *ServerUsersInterceptors) Authentified(ctx context.Context, info *users.AuthentifiedInfo, next goa.Endpoint) (any, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, users.MakeUnauthorized(fmt.Errorf("you need to be authenticated to access this endpoint"))
	}
	return next(ctx, info.RawPayload())
}

func (i *ServerUsersInterceptors) IsAdmin(ctx context.Context, info *users.IsAdminInfo, next goa.Endpoint) (any, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, users.MakeUnauthorized(fmt.Errorf("you need to be authenticated to access this endpoint"))
	}
	jwtClaims := token.(*service.JwtClaims)

	roleIds, err := i.userRolesRepository.GetUserRoleIDs(ctx, jwtClaims.UserID)
	if err != nil {
		return nil, users.MakeInternalError(fmt.Errorf("could not retrieve user roles: %w", err))
	}

	if len(roleIds) == 0 {

	}

	roles, err := i.rolesRepository.GetRolesByIDs(ctx, roleIds)

	if err != nil {
		return nil, users.MakeInternalError(fmt.Errorf("could not retrieve roles: %w", err))
	}
	for _, role := range roles {
		if role.Admin {
			return next(ctx, info.RawPayload())
		}
	}

	return nil, users.MakeForbidden(fmt.Errorf("you need to be an admin to access this endpoint"))
}

type ServerSecretsInterceptors struct {
	rolesRepository     repository.RolesRepository
	userRolesRepository repository.UserRolesRepository
}

func (i *ServerSecretsInterceptors) Authentified(ctx context.Context, info *secrets.AuthentifiedInfo, next goa.Endpoint) (any, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, secrets.MakeUnauthorized(fmt.Errorf("you need to be authenticated to access this endpoint"))
	}
	return next(ctx, info.RawPayload())
}
