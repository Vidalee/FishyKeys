// Code generated by goa v3.21.1, DO NOT EDIT.
//
// key_management service
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package keymanagement

import (
	"context"

	goa "goa.design/goa/v3/pkg"
)

// The FishyKeys server handles master key operations
type Service interface {
	// Create a new master key and split it into shares
	CreateMasterKey(context.Context, *CreateMasterKeyPayload) (res *CreateMasterKeyResult, err error)
	// Get the current status of the master key
	GetKeyStatus(context.Context) (res *GetKeyStatusResult, err error)
	// Add a share to unlock the master key
	AddShare(context.Context, *AddSharePayload) (res *AddShareResult, err error)
	// Delete a share from the key management system
	DeleteShare(context.Context, *DeleteSharePayload) (err error)
}

// APIName is the name of the API as defined in the design.
const APIName = "fishykeys"

// APIVersion is the version of the API as defined in the design.
const APIVersion = "1.0"

// ServiceName is the name of the service as defined in the design. This is the
// same value that is set in the endpoint request contexts under the ServiceKey
// key.
const ServiceName = "key_management"

// MethodNames lists the service method names as defined in the design. These
// are the same values that are set in the endpoint request contexts under the
// MethodKey key.
var MethodNames = [4]string{"create_master_key", "get_key_status", "add_share", "delete_share"}

// AddSharePayload is the payload type of the key_management service add_share
// method.
type AddSharePayload struct {
	// One of the shares need to unlock the master key
	Share string
}

// AddShareResult is the result type of the key_management service add_share
// method.
type AddShareResult struct {
	// The index of the share added
	Index int
	// Whether the master key has been unlocked
	Unlocked bool
}

// CreateMasterKeyPayload is the payload type of the key_management service
// create_master_key method.
type CreateMasterKeyPayload struct {
	// Total number of shares to create
	TotalShares int
	// Minimum number of shares required to reconstruct the key
	MinShares int
	// Admin username for key management
	AdminUsername string
	// Admin password for key management
	AdminPassword string
}

// CreateMasterKeyResult is the result type of the key_management service
// create_master_key method.
type CreateMasterKeyResult struct {
	// The generated key shares
	Shares []string
	// The admin user's username
	AdminUsername *string
}

// DeleteSharePayload is the payload type of the key_management service
// delete_share method.
type DeleteSharePayload struct {
	// The index of the share to delete
	Index int
}

// GetKeyStatusResult is the result type of the key_management service
// get_key_status method.
type GetKeyStatusResult struct {
	// Whether the key is currently locked
	IsLocked bool
	// Number of shares currently held
	CurrentShares int
	// Minimum number of shares required
	MinShares int
	// Total number of shares
	TotalShares int
}

// MakeInvalidParameters builds a goa.ServiceError from an error.
func MakeInvalidParameters(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "invalid_parameters", false, false, false)
}

// MakeInternalError builds a goa.ServiceError from an error.
func MakeInternalError(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "internal_error", false, false, false)
}

// MakeKeyAlreadyExists builds a goa.ServiceError from an error.
func MakeKeyAlreadyExists(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "key_already_exists", false, false, false)
}

// MakeNoKeySet builds a goa.ServiceError from an error.
func MakeNoKeySet(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "no_key_set", false, false, false)
}

// MakeTooManyShares builds a goa.ServiceError from an error.
func MakeTooManyShares(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "too_many_shares", false, false, false)
}

// MakeCouldNotRecombine builds a goa.ServiceError from an error.
func MakeCouldNotRecombine(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "could_not_recombine", false, false, false)
}

// MakeWrongShares builds a goa.ServiceError from an error.
func MakeWrongShares(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "wrong_shares", false, false, false)
}

// MakeKeyAlreadyUnlocked builds a goa.ServiceError from an error.
func MakeKeyAlreadyUnlocked(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "key_already_unlocked", false, false, false)
}

// MakeWrongIndex builds a goa.ServiceError from an error.
func MakeWrongIndex(err error) *goa.ServiceError {
	return goa.NewServiceError(err, "wrong_index", false, false, false)
}
