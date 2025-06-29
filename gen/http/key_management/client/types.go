// Code generated by goa v3.21.1, DO NOT EDIT.
//
// key_management HTTP client types
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package client

import (
	keymanagement "github.com/Vidalee/FishyKeys/gen/key_management"
	goa "goa.design/goa/v3/pkg"
)

// CreateMasterKeyRequestBody is the type of the "key_management" service
// "create_master_key" endpoint HTTP request body.
type CreateMasterKeyRequestBody struct {
	// Total number of shares to create
	TotalShares int `form:"total_shares" json:"total_shares" xml:"total_shares"`
	// Minimum number of shares required to reconstruct the key
	MinShares int `form:"min_shares" json:"min_shares" xml:"min_shares"`
	// Admin username for key management
	AdminUsername string `form:"admin_username" json:"admin_username" xml:"admin_username"`
	// Admin password for key management
	AdminPassword string `form:"admin_password" json:"admin_password" xml:"admin_password"`
}

// AddShareRequestBody is the type of the "key_management" service "add_share"
// endpoint HTTP request body.
type AddShareRequestBody struct {
	// One of the shares need to unlock the master key
	Share string `form:"share" json:"share" xml:"share"`
}

// DeleteShareRequestBody is the type of the "key_management" service
// "delete_share" endpoint HTTP request body.
type DeleteShareRequestBody struct {
	// The index of the share to delete
	Index int `form:"index" json:"index" xml:"index"`
}

// CreateMasterKeyResponseBody is the type of the "key_management" service
// "create_master_key" endpoint HTTP response body.
type CreateMasterKeyResponseBody struct {
	// The generated key shares
	Shares []string `form:"shares,omitempty" json:"shares,omitempty" xml:"shares,omitempty"`
	// The admin user's username
	AdminUsername *string `form:"admin_username,omitempty" json:"admin_username,omitempty" xml:"admin_username,omitempty"`
}

// GetKeyStatusResponseBody is the type of the "key_management" service
// "get_key_status" endpoint HTTP response body.
type GetKeyStatusResponseBody struct {
	// Whether the key is currently locked
	IsLocked *bool `form:"is_locked,omitempty" json:"is_locked,omitempty" xml:"is_locked,omitempty"`
	// Number of shares currently held
	CurrentShares *int `form:"current_shares,omitempty" json:"current_shares,omitempty" xml:"current_shares,omitempty"`
	// Minimum number of shares required
	MinShares *int `form:"min_shares,omitempty" json:"min_shares,omitempty" xml:"min_shares,omitempty"`
	// Total number of shares
	TotalShares *int `form:"total_shares,omitempty" json:"total_shares,omitempty" xml:"total_shares,omitempty"`
}

// AddShareResponseBody is the type of the "key_management" service "add_share"
// endpoint HTTP response body.
type AddShareResponseBody struct {
	// The index of the share added
	Index *int `form:"index,omitempty" json:"index,omitempty" xml:"index,omitempty"`
	// Whether the master key has been unlocked
	Unlocked *bool `form:"unlocked,omitempty" json:"unlocked,omitempty" xml:"unlocked,omitempty"`
}

// CreateMasterKeyInvalidParametersResponseBody is the type of the
// "key_management" service "create_master_key" endpoint HTTP response body for
// the "invalid_parameters" error.
type CreateMasterKeyInvalidParametersResponseBody struct {
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

// CreateMasterKeyInternalErrorResponseBody is the type of the "key_management"
// service "create_master_key" endpoint HTTP response body for the
// "internal_error" error.
type CreateMasterKeyInternalErrorResponseBody struct {
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

// NewCreateMasterKeyRequestBody builds the HTTP request body from the payload
// of the "create_master_key" endpoint of the "key_management" service.
func NewCreateMasterKeyRequestBody(p *keymanagement.CreateMasterKeyPayload) *CreateMasterKeyRequestBody {
	body := &CreateMasterKeyRequestBody{
		TotalShares:   p.TotalShares,
		MinShares:     p.MinShares,
		AdminUsername: p.AdminUsername,
		AdminPassword: p.AdminPassword,
	}
	return body
}

// NewAddShareRequestBody builds the HTTP request body from the payload of the
// "add_share" endpoint of the "key_management" service.
func NewAddShareRequestBody(p *keymanagement.AddSharePayload) *AddShareRequestBody {
	body := &AddShareRequestBody{
		Share: p.Share,
	}
	return body
}

// NewDeleteShareRequestBody builds the HTTP request body from the payload of
// the "delete_share" endpoint of the "key_management" service.
func NewDeleteShareRequestBody(p *keymanagement.DeleteSharePayload) *DeleteShareRequestBody {
	body := &DeleteShareRequestBody{
		Index: p.Index,
	}
	return body
}

// NewCreateMasterKeyResultCreated builds a "key_management" service
// "create_master_key" endpoint result from a HTTP "Created" response.
func NewCreateMasterKeyResultCreated(body *CreateMasterKeyResponseBody) *keymanagement.CreateMasterKeyResult {
	v := &keymanagement.CreateMasterKeyResult{
		AdminUsername: body.AdminUsername,
	}
	if body.Shares != nil {
		v.Shares = make([]string, len(body.Shares))
		for i, val := range body.Shares {
			v.Shares[i] = val
		}
	}

	return v
}

// NewCreateMasterKeyInvalidParameters builds a key_management service
// create_master_key endpoint invalid_parameters error.
func NewCreateMasterKeyInvalidParameters(body *CreateMasterKeyInvalidParametersResponseBody) *goa.ServiceError {
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

// NewCreateMasterKeyInternalError builds a key_management service
// create_master_key endpoint internal_error error.
func NewCreateMasterKeyInternalError(body *CreateMasterKeyInternalErrorResponseBody) *goa.ServiceError {
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

// NewCreateMasterKeyKeyAlreadyExists builds a key_management service
// create_master_key endpoint key_already_exists error.
func NewCreateMasterKeyKeyAlreadyExists(body string) keymanagement.KeyAlreadyExists {
	v := keymanagement.KeyAlreadyExists(body)

	return v
}

// NewGetKeyStatusResultOK builds a "key_management" service "get_key_status"
// endpoint result from a HTTP "OK" response.
func NewGetKeyStatusResultOK(body *GetKeyStatusResponseBody) *keymanagement.GetKeyStatusResult {
	v := &keymanagement.GetKeyStatusResult{
		IsLocked:      *body.IsLocked,
		CurrentShares: *body.CurrentShares,
		MinShares:     *body.MinShares,
		TotalShares:   *body.TotalShares,
	}

	return v
}

// NewGetKeyStatusInternalError builds a key_management service get_key_status
// endpoint internal_error error.
func NewGetKeyStatusInternalError(body string) keymanagement.InternalError {
	v := keymanagement.InternalError(body)

	return v
}

// NewGetKeyStatusNoKeySet builds a key_management service get_key_status
// endpoint no_key_set error.
func NewGetKeyStatusNoKeySet(body string) keymanagement.NoKeySet {
	v := keymanagement.NoKeySet(body)

	return v
}

// NewAddShareResultCreated builds a "key_management" service "add_share"
// endpoint result from a HTTP "Created" response.
func NewAddShareResultCreated(body *AddShareResponseBody) *keymanagement.AddShareResult {
	v := &keymanagement.AddShareResult{
		Index:    *body.Index,
		Unlocked: *body.Unlocked,
	}

	return v
}

// NewAddShareCouldNotRecombine builds a key_management service add_share
// endpoint could_not_recombine error.
func NewAddShareCouldNotRecombine(body string) keymanagement.CouldNotRecombine {
	v := keymanagement.CouldNotRecombine(body)

	return v
}

// NewAddShareInvalidParameters builds a key_management service add_share
// endpoint invalid_parameters error.
func NewAddShareInvalidParameters(body string) keymanagement.InvalidParameters {
	v := keymanagement.InvalidParameters(body)

	return v
}

// NewAddShareWrongShares builds a key_management service add_share endpoint
// wrong_shares error.
func NewAddShareWrongShares(body string) keymanagement.WrongShares {
	v := keymanagement.WrongShares(body)

	return v
}

// NewAddShareInternalError builds a key_management service add_share endpoint
// internal_error error.
func NewAddShareInternalError(body string) keymanagement.InternalError {
	v := keymanagement.InternalError(body)

	return v
}

// NewAddShareKeyAlreadyUnlocked builds a key_management service add_share
// endpoint key_already_unlocked error.
func NewAddShareKeyAlreadyUnlocked(body string) keymanagement.KeyAlreadyUnlocked {
	v := keymanagement.KeyAlreadyUnlocked(body)

	return v
}

// NewAddShareTooManyShares builds a key_management service add_share endpoint
// too_many_shares error.
func NewAddShareTooManyShares(body string) keymanagement.TooManyShares {
	v := keymanagement.TooManyShares(body)

	return v
}

// NewAddShareNoKeySet builds a key_management service add_share endpoint
// no_key_set error.
func NewAddShareNoKeySet(body string) keymanagement.NoKeySet {
	v := keymanagement.NoKeySet(body)

	return v
}

// NewDeleteShareInternalError builds a key_management service delete_share
// endpoint internal_error error.
func NewDeleteShareInternalError(body string) keymanagement.InternalError {
	v := keymanagement.InternalError(body)

	return v
}

// NewDeleteShareKeyAlreadyUnlocked builds a key_management service
// delete_share endpoint key_already_unlocked error.
func NewDeleteShareKeyAlreadyUnlocked(body string) keymanagement.KeyAlreadyUnlocked {
	v := keymanagement.KeyAlreadyUnlocked(body)

	return v
}

// NewDeleteShareNoKeySet builds a key_management service delete_share endpoint
// no_key_set error.
func NewDeleteShareNoKeySet(body string) keymanagement.NoKeySet {
	v := keymanagement.NoKeySet(body)

	return v
}

// NewDeleteShareWrongIndex builds a key_management service delete_share
// endpoint wrong_index error.
func NewDeleteShareWrongIndex(body string) keymanagement.WrongIndex {
	v := keymanagement.WrongIndex(body)

	return v
}

// ValidateGetKeyStatusResponseBody runs the validations defined on
// get_key_status_response_body
func ValidateGetKeyStatusResponseBody(body *GetKeyStatusResponseBody) (err error) {
	if body.IsLocked == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("is_locked", "body"))
	}
	if body.CurrentShares == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("current_shares", "body"))
	}
	if body.MinShares == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("min_shares", "body"))
	}
	if body.TotalShares == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("total_shares", "body"))
	}
	return
}

// ValidateAddShareResponseBody runs the validations defined on
// add_share_response_body
func ValidateAddShareResponseBody(body *AddShareResponseBody) (err error) {
	if body.Index == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("index", "body"))
	}
	if body.Unlocked == nil {
		err = goa.MergeErrors(err, goa.MissingFieldError("unlocked", "body"))
	}
	return
}

// ValidateCreateMasterKeyInvalidParametersResponseBody runs the validations
// defined on create_master_key_invalid_parameters_response_body
func ValidateCreateMasterKeyInvalidParametersResponseBody(body *CreateMasterKeyInvalidParametersResponseBody) (err error) {
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

// ValidateCreateMasterKeyInternalErrorResponseBody runs the validations
// defined on create_master_key_internal_error_response_body
func ValidateCreateMasterKeyInternalErrorResponseBody(body *CreateMasterKeyInternalErrorResponseBody) (err error) {
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
