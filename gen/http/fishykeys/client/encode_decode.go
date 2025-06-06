// Code generated by goa v3.21.1, DO NOT EDIT.
//
// fishykeys HTTP client encoders and decoders
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/backend/design

package client

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	fishykeys "github.com/Vidalee/FishyKeys/backend/gen/fishykeys"
	goahttp "goa.design/goa/v3/http"
)

// BuildCreateMasterKeyRequest instantiates a HTTP request object with method
// and path set to call the "fishykeys" service "create_master_key" endpoint
func (c *Client) BuildCreateMasterKeyRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: CreateMasterKeyFishykeysPath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("fishykeys", "create_master_key", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeCreateMasterKeyRequest returns an encoder for requests sent to the
// fishykeys create_master_key server.
func EncodeCreateMasterKeyRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
		p, ok := v.(*fishykeys.CreateMasterKeyPayload)
		if !ok {
			return goahttp.ErrInvalidType("fishykeys", "create_master_key", "*fishykeys.CreateMasterKeyPayload", v)
		}
		body := NewCreateMasterKeyRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("fishykeys", "create_master_key", err)
		}
		return nil
	}
}

// DecodeCreateMasterKeyResponse returns a decoder for responses returned by
// the fishykeys create_master_key endpoint. restoreBody controls whether the
// response body should be restored after having been read.
// DecodeCreateMasterKeyResponse may return the following errors:
//   - "internal_error" (type fishykeys.InternalError): http.StatusInternalServerError
//   - "invalid_parameters" (type fishykeys.InvalidParameters): http.StatusBadRequest
//   - "key_already_exists" (type fishykeys.KeyAlreadyExists): http.StatusConflict
//   - error: internal error
func DecodeCreateMasterKeyResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
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
		case http.StatusCreated:
			var (
				body CreateMasterKeyResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "create_master_key", err)
			}
			res := NewCreateMasterKeyResultCreated(&body)
			return res, nil
		case http.StatusInternalServerError:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "create_master_key", err)
			}
			return nil, NewCreateMasterKeyInternalError(body)
		case http.StatusBadRequest:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "create_master_key", err)
			}
			return nil, NewCreateMasterKeyInvalidParameters(body)
		case http.StatusConflict:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "create_master_key", err)
			}
			return nil, NewCreateMasterKeyKeyAlreadyExists(body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("fishykeys", "create_master_key", resp.StatusCode, string(body))
		}
	}
}

// BuildGetKeyStatusRequest instantiates a HTTP request object with method and
// path set to call the "fishykeys" service "get_key_status" endpoint
func (c *Client) BuildGetKeyStatusRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: GetKeyStatusFishykeysPath()}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("fishykeys", "get_key_status", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// DecodeGetKeyStatusResponse returns a decoder for responses returned by the
// fishykeys get_key_status endpoint. restoreBody controls whether the response
// body should be restored after having been read.
// DecodeGetKeyStatusResponse may return the following errors:
//   - "internal_error" (type fishykeys.InternalError): http.StatusInternalServerError
//   - "no_key_set" (type fishykeys.NoKeySet): http.StatusNotFound
//   - error: internal error
func DecodeGetKeyStatusResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
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
				body GetKeyStatusResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "get_key_status", err)
			}
			err = ValidateGetKeyStatusResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("fishykeys", "get_key_status", err)
			}
			res := NewGetKeyStatusResultOK(&body)
			return res, nil
		case http.StatusInternalServerError:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "get_key_status", err)
			}
			return nil, NewGetKeyStatusInternalError(body)
		case http.StatusNotFound:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "get_key_status", err)
			}
			return nil, NewGetKeyStatusNoKeySet(body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("fishykeys", "get_key_status", resp.StatusCode, string(body))
		}
	}
}

// BuildAddShareRequest instantiates a HTTP request object with method and path
// set to call the "fishykeys" service "add_share" endpoint
func (c *Client) BuildAddShareRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: AddShareFishykeysPath()}
	req, err := http.NewRequest("POST", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("fishykeys", "add_share", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeAddShareRequest returns an encoder for requests sent to the fishykeys
// add_share server.
func EncodeAddShareRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
		p, ok := v.(*fishykeys.AddSharePayload)
		if !ok {
			return goahttp.ErrInvalidType("fishykeys", "add_share", "*fishykeys.AddSharePayload", v)
		}
		body := NewAddShareRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("fishykeys", "add_share", err)
		}
		return nil
	}
}

// DecodeAddShareResponse returns a decoder for responses returned by the
// fishykeys add_share endpoint. restoreBody controls whether the response body
// should be restored after having been read.
// DecodeAddShareResponse may return the following errors:
//   - "could_not_recombine" (type fishykeys.CouldNotRecombine): http.StatusBadRequest
//   - "invalid_parameters" (type fishykeys.InvalidParameters): http.StatusBadRequest
//   - "wrong_shares" (type fishykeys.WrongShares): http.StatusBadRequest
//   - "internal_error" (type fishykeys.InternalError): http.StatusInternalServerError
//   - "key_already_unlocked" (type fishykeys.KeyAlreadyUnlocked): http.StatusConflict
//   - "too_many_shares" (type fishykeys.TooManyShares): http.StatusConflict
//   - "no_key_set" (type fishykeys.NoKeySet): http.StatusNotFound
//   - error: internal error
func DecodeAddShareResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
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
		case http.StatusCreated:
			var (
				body AddShareResponseBody
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
			}
			err = ValidateAddShareResponseBody(&body)
			if err != nil {
				return nil, goahttp.ErrValidationError("fishykeys", "add_share", err)
			}
			res := NewAddShareResultCreated(&body)
			return res, nil
		case http.StatusBadRequest:
			en := resp.Header.Get("goa-error")
			switch en {
			case "could_not_recombine":
				var (
					body string
					err  error
				)
				err = decoder(resp).Decode(&body)
				if err != nil {
					return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
				}
				return nil, NewAddShareCouldNotRecombine(body)
			case "invalid_parameters":
				var (
					body string
					err  error
				)
				err = decoder(resp).Decode(&body)
				if err != nil {
					return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
				}
				return nil, NewAddShareInvalidParameters(body)
			case "wrong_shares":
				var (
					body string
					err  error
				)
				err = decoder(resp).Decode(&body)
				if err != nil {
					return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
				}
				return nil, NewAddShareWrongShares(body)
			default:
				body, _ := io.ReadAll(resp.Body)
				return nil, goahttp.ErrInvalidResponse("fishykeys", "add_share", resp.StatusCode, string(body))
			}
		case http.StatusInternalServerError:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
			}
			return nil, NewAddShareInternalError(body)
		case http.StatusConflict:
			en := resp.Header.Get("goa-error")
			switch en {
			case "key_already_unlocked":
				var (
					body string
					err  error
				)
				err = decoder(resp).Decode(&body)
				if err != nil {
					return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
				}
				return nil, NewAddShareKeyAlreadyUnlocked(body)
			case "too_many_shares":
				var (
					body string
					err  error
				)
				err = decoder(resp).Decode(&body)
				if err != nil {
					return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
				}
				return nil, NewAddShareTooManyShares(body)
			default:
				body, _ := io.ReadAll(resp.Body)
				return nil, goahttp.ErrInvalidResponse("fishykeys", "add_share", resp.StatusCode, string(body))
			}
		case http.StatusNotFound:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "add_share", err)
			}
			return nil, NewAddShareNoKeySet(body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("fishykeys", "add_share", resp.StatusCode, string(body))
		}
	}
}

// BuildDeleteShareRequest instantiates a HTTP request object with method and
// path set to call the "fishykeys" service "delete_share" endpoint
func (c *Client) BuildDeleteShareRequest(ctx context.Context, v any) (*http.Request, error) {
	u := &url.URL{Scheme: c.scheme, Host: c.host, Path: DeleteShareFishykeysPath()}
	req, err := http.NewRequest("DELETE", u.String(), nil)
	if err != nil {
		return nil, goahttp.ErrInvalidURL("fishykeys", "delete_share", u.String(), err)
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}

	return req, nil
}

// EncodeDeleteShareRequest returns an encoder for requests sent to the
// fishykeys delete_share server.
func EncodeDeleteShareRequest(encoder func(*http.Request) goahttp.Encoder) func(*http.Request, any) error {
	return func(req *http.Request, v any) error {
		p, ok := v.(*fishykeys.DeleteSharePayload)
		if !ok {
			return goahttp.ErrInvalidType("fishykeys", "delete_share", "*fishykeys.DeleteSharePayload", v)
		}
		body := NewDeleteShareRequestBody(p)
		if err := encoder(req).Encode(&body); err != nil {
			return goahttp.ErrEncodingError("fishykeys", "delete_share", err)
		}
		return nil
	}
}

// DecodeDeleteShareResponse returns a decoder for responses returned by the
// fishykeys delete_share endpoint. restoreBody controls whether the response
// body should be restored after having been read.
// DecodeDeleteShareResponse may return the following errors:
//   - "internal_error" (type fishykeys.InternalError): http.StatusInternalServerError
//   - "key_already_unlocked" (type fishykeys.KeyAlreadyUnlocked): http.StatusConflict
//   - "no_key_set" (type fishykeys.NoKeySet): http.StatusNotFound
//   - "wrong_index" (type fishykeys.WrongIndex): http.StatusBadRequest
//   - error: internal error
func DecodeDeleteShareResponse(decoder func(*http.Response) goahttp.Decoder, restoreBody bool) func(*http.Response) (any, error) {
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
			return nil, nil
		case http.StatusInternalServerError:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "delete_share", err)
			}
			return nil, NewDeleteShareInternalError(body)
		case http.StatusConflict:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "delete_share", err)
			}
			return nil, NewDeleteShareKeyAlreadyUnlocked(body)
		case http.StatusNotFound:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "delete_share", err)
			}
			return nil, NewDeleteShareNoKeySet(body)
		case http.StatusBadRequest:
			var (
				body string
				err  error
			)
			err = decoder(resp).Decode(&body)
			if err != nil {
				return nil, goahttp.ErrDecodingError("fishykeys", "delete_share", err)
			}
			return nil, NewDeleteShareWrongIndex(body)
		default:
			body, _ := io.ReadAll(resp.Body)
			return nil, goahttp.ErrInvalidResponse("fishykeys", "delete_share", resp.StatusCode, string(body))
		}
	}
}
