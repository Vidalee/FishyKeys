package service

import (
	"context"
	"errors"
	"fmt"

	genroles "github.com/Vidalee/FishyKeys/gen/roles"
	"github.com/Vidalee/FishyKeys/repository"
)

type RolesService struct {
	rolesRepository     repository.RolesRepository
	userRepository      repository.UsersRepository
	userRolesRepository repository.UserRolesRepository
}

func NewRolesService(
	rolesRepository repository.RolesRepository,
	userRepository repository.UsersRepository,
	userRolesRepository repository.UserRolesRepository,
) *RolesService {
	return &RolesService{
		rolesRepository:     rolesRepository,
		userRepository:      userRepository,
		userRolesRepository: userRolesRepository,
	}
}

func (s *RolesService) ListRoles(ctx context.Context) ([]*genroles.Role, error) {
	roles, err := s.rolesRepository.ListRoles(ctx)
	if err != nil {
		return nil, genroles.MakeInternalError(fmt.Errorf("could not list roles: %w", err))
	}

	result := make([]*genroles.Role, 0, len(roles))
	for _, r := range roles {
		result = append(result, &genroles.Role{
			ID:        r.ID,
			Name:      r.Name,
			Color:     r.Color,
			Admin:     r.Admin,
			CreatedAt: r.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	return result, nil
}

func (s *RolesService) CreateRole(ctx context.Context, payload *genroles.CreateRolePayload) (*genroles.CreateRoleResult, error) {
	roles, err := s.rolesRepository.ListRoles(ctx)
	if err != nil {
		return nil, genroles.MakeInternalError(fmt.Errorf("could not list roles: %w", err))
	}
	for _, r := range roles {
		if r.Name == payload.Name {
			return nil, genroles.MakeRoleTaken(fmt.Errorf("role with name %s already exists", payload.Name))
		}
	}

	roleID, err := s.rolesRepository.CreateRole(ctx, payload.Name, payload.Color)
	if err != nil {
		return nil, genroles.MakeInternalError(fmt.Errorf("could not create role: %w", err))
	}

	return &genroles.CreateRoleResult{
		ID:    roleID,
		Color: payload.Color,
		Name:  payload.Name,
	}, nil
}

func (s *RolesService) DeleteRole(ctx context.Context, payload *genroles.DeleteRolePayload) error {
	if payload.ID == 1 {
		return genroles.MakeForbidden(fmt.Errorf("cannot delete admin role"))
	}

	err := s.rolesRepository.DeleteRole(ctx, payload.ID)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return genroles.MakeRoleNotFound(fmt.Errorf("role not found"))
		}
		return genroles.MakeInternalError(fmt.Errorf("could not delete role: %w", err))
	}
	return nil
}

func (s *RolesService) AssignRoleToUser(ctx context.Context, payload *genroles.AssignRoleToUserPayload) error {
	_, err := s.userRepository.GetUserByID(ctx, payload.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return genroles.MakeUserNotFound(fmt.Errorf("user not found"))
		}
		return genroles.MakeInternalError(fmt.Errorf("could not check user existence: %w", err))
	}

	_, err = s.rolesRepository.GetRoleByID(ctx, payload.RoleID)
	if err != nil {
		if errors.Is(err, repository.ErrRoleNotFound) {
			return genroles.MakeRoleNotFound(fmt.Errorf("role not found"))
		}
		return genroles.MakeInternalError(fmt.Errorf("could not check role existence: %w", err))
	}

	err = s.userRolesRepository.AssignRoleToUser(ctx, payload.UserID, payload.RoleID)
	if err != nil {
		return genroles.MakeInternalError(fmt.Errorf("could not assign role to user: %w", err))
	}

	return nil
}

func (s *RolesService) UnassignRoleToUser(ctx context.Context, payload *genroles.UnassignRoleToUserPayload) error {
	_, err := s.userRepository.GetUserByID(ctx, payload.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return genroles.MakeUserNotFound(fmt.Errorf("user not found"))
		}
		return genroles.MakeInternalError(fmt.Errorf("could not check user existence: %w", err))
	}

	roles, err := s.rolesRepository.ListRoles(ctx)
	if err != nil {
		return genroles.MakeInternalError(fmt.Errorf("could not list roles: %w", err))
	}

	var roleToUnassign *repository.Role
	for i := range roles {
		if roles[i].ID == payload.RoleID {
			roleToUnassign = &roles[i]
			break
		}
	}

	if roleToUnassign == nil {
		return genroles.MakeRoleNotFound(fmt.Errorf("role not found"))
	}

	if roleToUnassign.Admin {
		// Guaranteed by the IsAdmin interceptor
		if ctx.Value("token").(*JwtClaims).UserID == payload.UserID {
			return genroles.MakeForbidden(fmt.Errorf("cannot unassign admin role from yourself"))
		}
	}

	err = s.userRolesRepository.RemoveRoleFromUser(ctx, payload.UserID, payload.RoleID)
	if err != nil {
		return genroles.MakeInternalError(fmt.Errorf("could not unassign role from user: %w", err))
	}

	return nil
}
