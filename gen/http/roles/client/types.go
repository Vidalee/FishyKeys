// Code generated by goa v3.21.1, DO NOT EDIT.
//
// roles HTTP client types
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package client

import (
	roles "github.com/Vidalee/FishyKeys/gen/roles"
	goa "goa.design/goa/v3/pkg"
)

// ListRolesResponseBody is the type of the "roles" service "list roles"
// endpoint HTTP response body.
type ListRolesResponseBody []*RoleResponse

// ListRolesInternalErrorResponseBody is the type of the "roles" service "list
// roles" endpoint HTTP response body for the "internal_error" error.
type ListRolesInternalErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
	// Is the error temporary?
	Temporary *bool `form:"temporary,omitempty" json:"temporary,omitempty" xml:"temporary,omitempty"`
	// Is the error a timeout?
	Timeout *bool `form:"timeout,omitempty" json:"timeout,omitempty" xml:"timeout,omitempty"`
	// Is the error a server-side fault?
	Fault *bool `form:"fault,omitempty" json:"fault,omitempty" xml:"fault,omitempty"`
}

// ListRolesUnauthorizedResponseBody is the type of the "roles" service "list
// roles" endpoint HTTP response body for the "unauthorized" error.
type ListRolesUnauthorizedResponseBody struct {
	// Name is the name of this class of errors.
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID *string `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message *string `form:"message,omitempty" json:"message,omitempty" xml:"message,omitempty"`
	// Is the error temporary?
	Temporary *bool `form:"temporary,omitempty" json:"temporary,omitempty" xml:"temporary,omitempty"`
	// Is the error a timeout?
	Timeout *bool `form:"timeout,omitempty" json:"timeout,omitempty" xml:"timeout,omitempty"`
	// Is the error a server-side fault?
	Fault *bool `form:"fault,omitempty" json:"fault,omitempty" xml:"fault,omitempty"`
}

// RoleResponse is used to define fields on response body types.
type RoleResponse struct {
	// Unique identifier for the role
	ID *int `form:"id,omitempty" json:"id,omitempty" xml:"id,omitempty"`
	// Name of the role
	Name *string `form:"name,omitempty" json:"name,omitempty" xml:"name,omitempty"`
	// Color associated with the role
	Color *string `form:"color,omitempty" json:"color,omitempty" xml:"color,omitempty"`
	// Is this role an admin role?
	Admin *bool `form:"admin,omitempty" json:"admin,omitempty" xml:"admin,omitempty"`
	// Role creation timestamp
	CreatedAt *string `form:"created_at,omitempty" json:"created_at,omitempty" xml:"created_at,omitempty"`
	// Role last update timestamp
	UpdatedAt *string `form:"updated_at,omitempty" json:"updated_at,omitempty" xml:"updated_at,omitempty"`
}

// NewListRolesRoleOK builds a "roles" service "list roles" endpoint result
// from a HTTP "OK" response.
func NewListRolesRoleOK(body []*RoleResponse) []*roles.Role {
	v := make([]*roles.Role, len(body))
	for i, val := range body {
		v[i] = unmarshalRoleResponseToRolesRole(val)
	}

	return v
}

// NewListRolesInternalError builds a roles service list roles endpoint
// internal_error error.
func NewListRolesInternalError(body *ListRolesInternalErrorResponseBody) *goa.ServiceError {
	v := &goa.ServiceError{
		Name:      *body.Name,
		ID:        *body.ID,
		Message:   *body.Message,
		Temporary: *body.Temporary,
		Timeout:   *body.Timeout,
		Fault:     *body.Fault,
	}

	return v
}

// NewListRolesUnauthorized builds a roles service list roles endpoint
// unauthorized error.
func NewListRolesUnauthorized(body *ListRolesUnauthorizedResponseBody) *goa.ServiceError {
	v := &goa.ServiceError{
		Name:      *body.Name,
		ID:        *body.ID,
		Message:   *body.Message,
		Temporary: *body.Temporary,
		Timeout:   *body.Timeout,
		Fault:     *body.Fault,
	}

	return v
}

// ValidateListRolesInternalErrorResponseBody runs the validations defined on
// list roles_internal_error_response_body
func ValidateListRolesInternalErrorResponseBody(body *ListRolesInternalErrorResponseBody) (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Message == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("message", "body"))
	}
	if body.Temporary == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("temporary", "body"))
	}
	if body.Timeout == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("timeout", "body"))
	}
	if body.Fault == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("fault", "body"))
	}
	return
}

// ValidateListRolesUnauthorizedResponseBody runs the validations defined on
// list roles_unauthorized_response_body
func ValidateListRolesUnauthorizedResponseBody(body *ListRolesUnauthorizedResponseBody) (err error) {
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Message == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("message", "body"))
	}
	if body.Temporary == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("temporary", "body"))
	}
	if body.Timeout == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("timeout", "body"))
	}
	if body.Fault == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("fault", "body"))
	}
	return
}

// ValidateRoleResponse runs the validations defined on RoleResponse
func ValidateRoleResponse(body *RoleResponse) (err error) {
	if body.ID == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("id", "body"))
	}
	if body.Name == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("name", "body"))
	}
	if body.Color == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("color", "body"))
	}
	if body.Admin == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("admin", "body"))
	}
	if body.CreatedAt == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("created_at", "body"))
	}
	if body.UpdatedAt == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("updated_at", "body"))
	}
	return
}
