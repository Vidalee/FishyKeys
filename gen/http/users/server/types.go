// Code generated by goa v3.21.1, DO NOT EDIT.
//
// users HTTP server types
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package server

import (
	"unicode/utf8"

	users "github.com/Vidalee/FishyKeys/gen/users"
	goa "goa.design/goa/v3/pkg"
)

// CreateUserRequestBody is the type of the "users" service "create user"
// endpoint HTTP request body.
type CreateUserRequestBody struct {
	// Username of the new user
	Username *string `form:"username,omitempty" json:"username,omitempty" xml:"username,omitempty"`
	// Password (hashed or plain depending on implementation)
	Password *string `form:"password,omitempty" json:"password,omitempty" xml:"password,omitempty"`
}

// AuthUserRequestBody is the type of the "users" service "auth user" endpoint
// HTTP request body.
type AuthUserRequestBody struct {
	// Username
	Username *string `form:"username,omitempty" json:"username,omitempty" xml:"username,omitempty"`
	// Password
	Password *string `form:"password,omitempty" json:"password,omitempty" xml:"password,omitempty"`
}

// CreateUserResponseBody is the type of the "users" service "create user"
// endpoint HTTP response body.
type CreateUserResponseBody struct {
	// The username of the created user
	Username string `form:"username" json:"username" xml:"username"`
	// Unique identifier for the user
	ID int `form:"id" json:"id" xml:"id"`
}

// ListUsersResponseBody is the type of the "users" service "list users"
// endpoint HTTP response body.
type ListUsersResponseBody []*UserResponse

// AuthUserResponseBody is the type of the "users" service "auth user" endpoint
// HTTP response body.
type AuthUserResponseBody struct {
	// The username of the authenticated user
	Username *string `form:"username,omitempty" json:"username,omitempty" xml:"username,omitempty"`
	// JWT or session token
	Token *string `form:"token,omitempty" json:"token,omitempty" xml:"token,omitempty"`
}

// GetOperatorTokenResponseBody is the type of the "users" service "get
// operator token" endpoint HTTP response body.
type GetOperatorTokenResponseBody struct {
	// The username of the account corresponding to the token
	Username *string `form:"username,omitempty" json:"username,omitempty" xml:"username,omitempty"`
	// JWT or session token
	Token *string `form:"token,omitempty" json:"token,omitempty" xml:"token,omitempty"`
}

// CreateUserUsernameTakenResponseBody is the type of the "users" service
// "create user" endpoint HTTP response body for the "username_taken" error.
type CreateUserUsernameTakenResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// CreateUserInvalidParametersResponseBody is the type of the "users" service
// "create user" endpoint HTTP response body for the "invalid_parameters" error.
type CreateUserInvalidParametersResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// CreateUserInternalErrorResponseBody is the type of the "users" service
// "create user" endpoint HTTP response body for the "internal_error" error.
type CreateUserInternalErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ListUsersInternalErrorResponseBody is the type of the "users" service "list
// users" endpoint HTTP response body for the "internal_error" error.
type ListUsersInternalErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// ListUsersUnauthorizedResponseBody is the type of the "users" service "list
// users" endpoint HTTP response body for the "unauthorized" error.
type ListUsersUnauthorizedResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// DeleteUserUserNotFoundResponseBody is the type of the "users" service
// "delete user" endpoint HTTP response body for the "user_not_found" error.
type DeleteUserUserNotFoundResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// DeleteUserInternalErrorResponseBody is the type of the "users" service
// "delete user" endpoint HTTP response body for the "internal_error" error.
type DeleteUserInternalErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// DeleteUserInvalidParametersResponseBody is the type of the "users" service
// "delete user" endpoint HTTP response body for the "invalid_parameters" error.
type DeleteUserInvalidParametersResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// DeleteUserForbiddenResponseBody is the type of the "users" service "delete
// user" endpoint HTTP response body for the "forbidden" error.
type DeleteUserForbiddenResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// DeleteUserUnauthorizedResponseBody is the type of the "users" service
// "delete user" endpoint HTTP response body for the "unauthorized" error.
type DeleteUserUnauthorizedResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// AuthUserUnauthorizedResponseBody is the type of the "users" service "auth
// user" endpoint HTTP response body for the "unauthorized" error.
type AuthUserUnauthorizedResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// AuthUserInvalidParametersResponseBody is the type of the "users" service
// "auth user" endpoint HTTP response body for the "invalid_parameters" error.
type AuthUserInvalidParametersResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// AuthUserInternalErrorResponseBody is the type of the "users" service "auth
// user" endpoint HTTP response body for the "internal_error" error.
type AuthUserInternalErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// GetOperatorTokenInternalErrorResponseBody is the type of the "users" service
// "get operator token" endpoint HTTP response body for the "internal_error"
// error.
type GetOperatorTokenInternalErrorResponseBody struct {
	// Name is the name of this class of errors.
	Name string `form:"name" json:"name" xml:"name"`
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `form:"id" json:"id" xml:"id"`
	// Message is a human-readable explanation specific to this occurrence of the
	// problem.
	Message string `form:"message" json:"message" xml:"message"`
	// Is the error temporary?
	Temporary bool `form:"temporary" json:"temporary" xml:"temporary"`
	// Is the error a timeout?
	Timeout bool `form:"timeout" json:"timeout" xml:"timeout"`
	// Is the error a server-side fault?
	Fault bool `form:"fault" json:"fault" xml:"fault"`
}

// UserResponse is used to define fields on response body types.
type UserResponse struct {
	// Unique identifier for the user
	ID int `form:"id" json:"id" xml:"id"`
	// The username
	Username string `form:"username" json:"username" xml:"username"`
	// User creation timestamp
	CreatedAt string `form:"created_at" json:"created_at" xml:"created_at"`
	// User last update timestamp
	UpdatedAt string `form:"updated_at" json:"updated_at" xml:"updated_at"`
}

// NewCreateUserResponseBody builds the HTTP response body from the result of
// the "create user" endpoint of the "users" service.
func NewCreateUserResponseBody(res *users.CreateUserResult) *CreateUserResponseBody {
	body := &CreateUserResponseBody{
		Username: res.Username,
		ID:       res.ID,
	}
	return body
}

// NewListUsersResponseBody builds the HTTP response body from the result of
// the "list users" endpoint of the "users" service.
func NewListUsersResponseBody(res []*users.User) ListUsersResponseBody {
	body := make([]*UserResponse, len(res))
	for i, val := range res {
		body[i] = marshalUsersUserToUserResponse(val)
	}
	return body
}

// NewAuthUserResponseBody builds the HTTP response body from the result of the
// "auth user" endpoint of the "users" service.
func NewAuthUserResponseBody(res *users.AuthUserResult) *AuthUserResponseBody {
	body := &AuthUserResponseBody{
		Username: res.Username,
		Token:    res.Token,
	}
	return body
}

// NewGetOperatorTokenResponseBody builds the HTTP response body from the
// result of the "get operator token" endpoint of the "users" service.
func NewGetOperatorTokenResponseBody(res *users.GetOperatorTokenResult) *GetOperatorTokenResponseBody {
	body := &GetOperatorTokenResponseBody{
		Username: res.Username,
		Token:    res.Token,
	}
	return body
}

// NewCreateUserUsernameTakenResponseBody builds the HTTP response body from
// the result of the "create user" endpoint of the "users" service.
func NewCreateUserUsernameTakenResponseBody(res *goa.ServiceError) *CreateUserUsernameTakenResponseBody {
	body := &CreateUserUsernameTakenResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewCreateUserInvalidParametersResponseBody builds the HTTP response body
// from the result of the "create user" endpoint of the "users" service.
func NewCreateUserInvalidParametersResponseBody(res *goa.ServiceError) *CreateUserInvalidParametersResponseBody {
	body := &CreateUserInvalidParametersResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewCreateUserInternalErrorResponseBody builds the HTTP response body from
// the result of the "create user" endpoint of the "users" service.
func NewCreateUserInternalErrorResponseBody(res *goa.ServiceError) *CreateUserInternalErrorResponseBody {
	body := &CreateUserInternalErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewListUsersInternalErrorResponseBody builds the HTTP response body from the
// result of the "list users" endpoint of the "users" service.
func NewListUsersInternalErrorResponseBody(res *goa.ServiceError) *ListUsersInternalErrorResponseBody {
	body := &ListUsersInternalErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewListUsersUnauthorizedResponseBody builds the HTTP response body from the
// result of the "list users" endpoint of the "users" service.
func NewListUsersUnauthorizedResponseBody(res *goa.ServiceError) *ListUsersUnauthorizedResponseBody {
	body := &ListUsersUnauthorizedResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewDeleteUserUserNotFoundResponseBody builds the HTTP response body from the
// result of the "delete user" endpoint of the "users" service.
func NewDeleteUserUserNotFoundResponseBody(res *goa.ServiceError) *DeleteUserUserNotFoundResponseBody {
	body := &DeleteUserUserNotFoundResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewDeleteUserInternalErrorResponseBody builds the HTTP response body from
// the result of the "delete user" endpoint of the "users" service.
func NewDeleteUserInternalErrorResponseBody(res *goa.ServiceError) *DeleteUserInternalErrorResponseBody {
	body := &DeleteUserInternalErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewDeleteUserInvalidParametersResponseBody builds the HTTP response body
// from the result of the "delete user" endpoint of the "users" service.
func NewDeleteUserInvalidParametersResponseBody(res *goa.ServiceError) *DeleteUserInvalidParametersResponseBody {
	body := &DeleteUserInvalidParametersResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewDeleteUserForbiddenResponseBody builds the HTTP response body from the
// result of the "delete user" endpoint of the "users" service.
func NewDeleteUserForbiddenResponseBody(res *goa.ServiceError) *DeleteUserForbiddenResponseBody {
	body := &DeleteUserForbiddenResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewDeleteUserUnauthorizedResponseBody builds the HTTP response body from the
// result of the "delete user" endpoint of the "users" service.
func NewDeleteUserUnauthorizedResponseBody(res *goa.ServiceError) *DeleteUserUnauthorizedResponseBody {
	body := &DeleteUserUnauthorizedResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewAuthUserUnauthorizedResponseBody builds the HTTP response body from the
// result of the "auth user" endpoint of the "users" service.
func NewAuthUserUnauthorizedResponseBody(res *goa.ServiceError) *AuthUserUnauthorizedResponseBody {
	body := &AuthUserUnauthorizedResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewAuthUserInvalidParametersResponseBody builds the HTTP response body from
// the result of the "auth user" endpoint of the "users" service.
func NewAuthUserInvalidParametersResponseBody(res *goa.ServiceError) *AuthUserInvalidParametersResponseBody {
	body := &AuthUserInvalidParametersResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewAuthUserInternalErrorResponseBody builds the HTTP response body from the
// result of the "auth user" endpoint of the "users" service.
func NewAuthUserInternalErrorResponseBody(res *goa.ServiceError) *AuthUserInternalErrorResponseBody {
	body := &AuthUserInternalErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewGetOperatorTokenInternalErrorResponseBody builds the HTTP response body
// from the result of the "get operator token" endpoint of the "users" service.
func NewGetOperatorTokenInternalErrorResponseBody(res *goa.ServiceError) *GetOperatorTokenInternalErrorResponseBody {
	body := &GetOperatorTokenInternalErrorResponseBody{
		Name:      res.Name,
		ID:        res.ID,
		Message:   res.Message,
		Temporary: res.Temporary,
		Timeout:   res.Timeout,
		Fault:     res.Fault,
	}
	return body
}

// NewCreateUserPayload builds a users service create user endpoint payload.
func NewCreateUserPayload(body *CreateUserRequestBody) *users.CreateUserPayload {
	v := &users.CreateUserPayload{
		Username: *body.Username,
		Password: *body.Password,
	}

	return v
}

// NewDeleteUserPayload builds a users service delete user endpoint payload.
func NewDeleteUserPayload(username string) *users.DeleteUserPayload {
	v := &users.DeleteUserPayload{}
	v.Username = username

	return v
}

// NewAuthUserPayload builds a users service auth user endpoint payload.
func NewAuthUserPayload(body *AuthUserRequestBody) *users.AuthUserPayload {
	v := &users.AuthUserPayload{
		Username: *body.Username,
		Password: *body.Password,
	}

	return v
}

// ValidateCreateUserRequestBody runs the validations defined on Create
// UserRequestBody
func ValidateCreateUserRequestBody(body *CreateUserRequestBody) (err error) {
	if body.Username == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("username", "body"))
	}
	if body.Password == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("password", "body"))
	}
	if body.Username != nil {
		if utf8.RuneCountInString(*body.Username) < 3 {
			err = goa.MergeErrors(err, goa.InvalidLengthError("body.username", *body.Username, utf8.RuneCountInString(*body.Username), 3, true))
		}
	}
	return
}

// ValidateAuthUserRequestBody runs the validations defined on Auth
// UserRequestBody
func ValidateAuthUserRequestBody(body *AuthUserRequestBody) (err error) {
	if body.Username == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("username", "body"))
	}
	if body.Password == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("password", "body"))
	}
	return
}
