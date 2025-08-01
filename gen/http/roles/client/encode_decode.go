// Code generated by goa v3.21.1, DO NOT EDIT.
//
// roles HTTP client encoders and decoders
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/design

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	roles "github.com/Vidalee/FishyKeys/gen/roles"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// BuildListRolesRequest instantiates a HTTP request object with method and
// path set to call the "roles" service "list roles" endpoint
func (c *Client) BuildListRolesRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: ListRolesRolesPath()}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("roles", "list roles", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeListRolesResponse returns a decoder for responses returned by the
// roles list roles endpoint. restoreBody controls whether the response body
// should be restored after having been read.
// DecodeListRolesResponse may return the following errors:
//   - "internal_error" (type *goa.ServiceError): http.StatusInternalServerError
//   - "unauthorized" (type *goa.ServiceError): http.StatusUnauthorized
//   - error: internal error
func DecodeListRolesResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
	return func(resp *http.Response) (any, error) {
		if restoreBody {
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
			resp.Body = io.NopCloser(bytes.NewBuffer(b))
			defer func() {
				resp.Body = io.NopCloser(bytes.NewBuffer(b))
			}()
		} else {
			defer resp.Body.Close()
		}
		switch resp.StatusCode {
		case http.StatusOK:
			var (
				body ListRolesResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("roles", "list roles", err)
			}
			for _, e := range body {
				if e != nil {
					if err2 := ValidateRoleResponse(e); err2 != nil {
						err = goa.MergeErrors(err, err2)
					}
				}
			}
			if err != nil {
				return nil, goahttp.ErrValidationError("roles", "list roles", err)
			}
			res := NewListRolesRoleOK(body)
			return res, nil
		case http.StatusInternalServerError:
			var (
				body ListRolesInternalErrorResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("roles", "list roles", err)
			}
			err = ValidateListRolesInternalErrorResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("roles", "list roles", err)
			}
			return nil, NewListRolesInternalError(&body)
		case http.StatusUnauthorized:
			var (
				body ListRolesUnauthorizedResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("roles", "list roles", err)
			}
			err = ValidateListRolesUnauthorizedResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("roles", "list roles", err)
			}
			return nil, NewListRolesUnauthorized(&body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("roles", "list roles", resp.StatusCode, string(body))
		}
	}
}

// unmarshalRoleResponseToRolesRole builds a value of type *roles.Role from a
// value of type *RoleResponse.
func unmarshalRoleResponseToRolesRole(v *RoleResponse) *roles.Role {
	res := &roles.Role{
		ID:        *v.ID,
		Name:      *v.Name,
		Color:     *v.Color,
		Admin:     *v.Admin,
		CreatedAt: *v.CreatedAt,
		UpdatedAt: *v.UpdatedAt,
	}

	return res
}
