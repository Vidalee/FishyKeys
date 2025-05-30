// Code generated by goa v3.21.1, DO NOT EDIT.
//
// fishykeys HTTP server encoders and decoders
//
// Command:
// $ goa gen github.com/Vidalee/FishyKeys/backend/design

package server

import (
	"context"
	"errors"
	"io"
	"net/http"

	fishykeys "github.com/Vidalee/FishyKeys/backend/gen/fishykeys"
	goahttp "goa.design/goa/v3/http"
	goa "goa.design/goa/v3/pkg"
)

// EncodeCreateMasterKeyResponse returns an encoder for responses returned by
// the fishykeys create_master_key endpoint.
func EncodeCreateMasterKeyResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res, _ := v.(*fishykeys.CreateMasterKeyResult)
		enc := encoder(ctx, w)
		body := NewCreateMasterKeyResponseBody(res)
		w.WriteHeader(http.StatusCreated)
		return enc.Encode(body)
	}
}

// DecodeCreateMasterKeyRequest returns a decoder for requests sent to the
// fishykeys create_master_key endpoint.
func DecodeCreateMasterKeyRequest(mux goahttp.Muxer, decoder func(*http.Request) goahttp.Decoder) func(*http.Request) (any, error) {
	return func(r *http.Request) (any, error) {
		var (
			body CreateMasterKeyRequestBody
			err  error
		)
		err = decoder(r).Decode(&body)
		if err != nil {
			if err == io.EOF {
				return nil, goa.MissingPayloadError()
			}
			var gerr *goa.ServiceError
			if errors.As(err, &gerr) {
				return nil, gerr
			}
			return nil, goa.DecodePayloadError(err.Error())
		}
		err = ValidateCreateMasterKeyRequestBody(&body)
		if err != nil {
			return nil, err
		}
		payload := NewCreateMasterKeyPayload(&body)

		return payload, nil
	}
}

// EncodeCreateMasterKeyError returns an encoder for errors returned by the
// create_master_key fishykeys endpoint.
func EncodeCreateMasterKeyError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(ctx context.Context, err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en goa.GoaErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.GoaErrorName() {
		case "internal_error":
			var res fishykeys.InternalError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.GoaErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		case "invalid_parameters":
			var res fishykeys.InvalidParameters
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.GoaErrorName())
			w.WriteHeader(http.StatusBadRequest)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}

// EncodeGetKeyStatusResponse returns an encoder for responses returned by the
// fishykeys get_key_status endpoint.
func EncodeGetKeyStatusResponse(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder) func(context.Context, http.ResponseWriter, any) error {
	return func(ctx context.Context, w http.ResponseWriter, v any) error {
		res, _ := v.(*fishykeys.GetKeyStatusResult)
		enc := encoder(ctx, w)
		body := NewGetKeyStatusResponseBody(res)
		w.WriteHeader(http.StatusOK)
		return enc.Encode(body)
	}
}

// EncodeGetKeyStatusError returns an encoder for errors returned by the
// get_key_status fishykeys endpoint.
func EncodeGetKeyStatusError(encoder func(context.Context, http.ResponseWriter) goahttp.Encoder, formatter func(ctx context.Context, err error) goahttp.Statuser) func(context.Context, http.ResponseWriter, error) error {
	encodeError := goahttp.ErrorEncoder(encoder, formatter)
	return func(ctx context.Context, w http.ResponseWriter, v error) error {
		var en goa.GoaErrorNamer
		if !errors.As(v, &en) {
			return encodeError(ctx, w, v)
		}
		switch en.GoaErrorName() {
		case "internal_error":
			var res fishykeys.InternalError
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.GoaErrorName())
			w.WriteHeader(http.StatusInternalServerError)
			return enc.Encode(body)
		case "no_key_set":
			var res fishykeys.NoKeySet
			errors.As(v, &res)
			enc := encoder(ctx, w)
			body := res
			w.Header().Set("goa-error", res.GoaErrorName())
			w.WriteHeader(http.StatusNotFound)
			return enc.Encode(body)
		default:
			return encodeError(ctx, w, v)
		}
	}
}
