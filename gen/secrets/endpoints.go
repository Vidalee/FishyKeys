// Code generated by goa v3.21.1, DO NOT EDIT.
//
// secrets endpoints
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package secrets

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// Endpoints wraps the "secrets" service endpoints.
type Endpoints struct {
	ListSecrets            goa.Endpoint
	GetSecretValue         goa.Endpoint
	OperatorGetSecretValue goa.Endpoint
	GetSecret              goa.Endpoint
	CreateSecret           goa.Endpoint
}

// NewEndpoints wraps the methods of the "secrets" service with endpoints.
func NewEndpoints(s Service, si ServerInterceptors) *Endpoints {
	endpoints := &Endpoints{
		ListSecrets:            NewListSecretsEndpoint(s),
		GetSecretValue:         NewGetSecretValueEndpoint(s),
		OperatorGetSecretValue: NewOperatorGetSecretValueEndpoint(s),
		GetSecret:              NewGetSecretEndpoint(s),
		CreateSecret:           NewCreateSecretEndpoint(s),
	}
	endpoints.ListSecrets = WrapListSecretsEndpoint(endpoints.ListSecrets, si)
	endpoints.GetSecretValue = WrapGetSecretValueEndpoint(endpoints.GetSecretValue, si)
	endpoints.GetSecret = WrapGetSecretEndpoint(endpoints.GetSecret, si)
	endpoints.CreateSecret = WrapCreateSecretEndpoint(endpoints.CreateSecret, si)
	return endpoints
}

// Use applies the given middleware to all the "secrets" service endpoints.
func (e *Endpoints) Use(m func(goa.Endpoint) goa.Endpoint) {
	e.ListSecrets = m(e.ListSecrets)
	e.GetSecretValue = m(e.GetSecretValue)
	e.OperatorGetSecretValue = m(e.OperatorGetSecretValue)
	e.GetSecret = m(e.GetSecret)
	e.CreateSecret = m(e.CreateSecret)
}

// NewListSecretsEndpoint returns an endpoint function that calls the method
// "list secrets" of service "secrets".
func NewListSecretsEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		return s.ListSecrets(ctx)
	}
}

// NewGetSecretValueEndpoint returns an endpoint function that calls the method
// "get secret value" of service "secrets".
func NewGetSecretValueEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*GetSecretValuePayload)
		return s.GetSecretValue(ctx, p)
	}
}

// NewOperatorGetSecretValueEndpoint returns an endpoint function that calls
// the method "operator get secret value" of service "secrets".
func NewOperatorGetSecretValueEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*OperatorGetSecretValuePayload)
		return s.OperatorGetSecretValue(ctx, p)
	}
}

// NewGetSecretEndpoint returns an endpoint function that calls the method "get
// secret" of service "secrets".
func NewGetSecretEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*GetSecretPayload)
		return s.GetSecret(ctx, p)
	}
}

// NewCreateSecretEndpoint returns an endpoint function that calls the method
// "create secret" of service "secrets".
func NewCreateSecretEndpoint(s Service) goa.Endpoint {
	return func(ctx context.Context, req any) (any, error) {
		p := req.(*CreateSecretPayload)
		return nil, s.CreateSecret(ctx, p)
	}
}
