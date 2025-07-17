package service

import (
	"context"
	"fmt"
	genroles "github.com/Vidalee/FishyKeys/gen/roles"
	genusers "github.com/Vidalee/FishyKeys/gen/users"
	"github.com/Vidalee/FishyKeys/repository"
)

type RolesService struct {
	rolesRepository repository.RolesRepository
}

func NewRolesService(
	rolesRepository repository.RolesRepository,
) *RolesService {
	return &RolesService{
		rolesRepository: rolesRepository,
	}
}

func (s *RolesService) ListRoles(ctx context.Context) ([]*genroles.Role, error) {
	roles, err := s.rolesRepository.ListRoles(ctx)
	if err != nil {
		return nil, genusers.MakeInternalError(fmt.Errorf("could not list users: %w", err))
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
