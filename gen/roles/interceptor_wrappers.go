// Code generated by goa v3.21.1, DO NOT EDIT.
//
// Interceptor wrappers
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package roles

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// wrapAuthentifiedListRoles applies the Authentified server interceptor to
// endpoints.
func wrapListRolesAuthentified(endpoint goa.Endpoint, i ServerInterceptors) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		info := &AuthentifiedInfo{
			service:    "roles",
			method:     "ListRoles",
			callType:   goa.InterceptorUnary,
			rawPayload: req,
		}
		return i.Authentified(ctx, info, endpoint)
	}
}
