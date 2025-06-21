package server

import (
	"context"
	"fmt"
	"github.com/Vidalee/FishyKeys/gen/users"
	goa "goa.design/goa/v3/pkg"
)

type ServerInterceptors struct{}

func (i *ServerInterceptors) Authentified(ctx context.Context, info *users.AuthentifiedInfo, next goa.Endpoint) (any, error) {
	token := ctx.Value("token")
	if token == nil {
		return nil, users.MakeUnauthorized(fmt.Errorf("you need to be authenticated to access this endpoint"))
	}
	return next(ctx, info.RawPayload())
}
