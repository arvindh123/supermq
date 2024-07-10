// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pats

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/absmach/magistrala/auth"
	"github.com/absmach/magistrala/internal/api"
	"github.com/absmach/magistrala/internal/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	"github.com/go-chi/chi/v5"
	kithttp "github.com/go-kit/kit/transport/http"
)

const (
	contentType = "application/json"
	defInterval = "30d"
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc auth.Service, mux *chi.Mux, logger *slog.Logger) *chi.Mux {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(apiutil.LoggingErrorEncoder(logger, api.EncodeError)),
	}
	mux.Route("/pats", func(r chi.Router) {
		r.Post("/", kithttp.NewServer(
			createPatEndpoint(svc),
			decodeCreatePatRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Get("/{id}", kithttp.NewServer(
			(retrieveEndpoint(svc)),
			decodeRetrievePatRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Put("/{id}/name", kithttp.NewServer(
			(updateNameEndpoint(svc)),
			decodeUpdateNameRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Put("/{id}/description", kithttp.NewServer(
			(updateDescriptionEndpoint(svc)),
			decodeUpdateDescriptionRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Get("/", kithttp.NewServer(
			(listEndpoint(svc)),
			decodeListRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Delete("/{id}", kithttp.NewServer(
			(deleteEndpoint(svc)),
			decodeDeleteRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Put("/{id}/secret/reset", kithttp.NewServer(
			(resetSecretEndpoint(svc)),
			decodeResetSecretRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Put("/{id}/secret/revoke", kithttp.NewServer(
			(revokeSecretEndpoint(svc)),
			decodeRevokeSecretRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Put("/{id}/scope/add", kithttp.NewServer(
			(addScopeEntryEndpoint(svc)),
			decodeAddScopeEntryRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Put("/{id}/scope/remove", kithttp.NewServer(
			(removeScopeEntryEndpoint(svc)),
			decodeRemoveScopeEntryRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Delete("/{id}/scope", kithttp.NewServer(
			(clearAllScopeEntryEndpoint(svc)),
			decodeClearAllScopeEntryRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)

		r.Get("/check", kithttp.NewServer(
			(testCheckScopeEntryEndpoint(svc)),
			decodeTestCheckScopeEntryRequest,
			api.EncodeResponse,
			opts...,
		).ServeHTTP)
	})
	return mux
}

func decodeCreatePatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := createPatReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, errors.Wrap(err, errors.ErrMalformedEntity))
	}
	return req, nil
}

func decodeRetrievePatRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := retrievePatReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	return req, nil
}

func decodeUpdateNameRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updatePatNameReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}

func decodeUpdateDescriptionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := updatePatDescriptionReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}

func decodeListRequest(_ context.Context, r *http.Request) (interface{}, error) {
	l, err := apiutil.ReadNumQuery[uint64](r, api.LimitKey, api.DefLimit)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	o, err := apiutil.ReadNumQuery[uint64](r, api.OffsetKey, api.DefOffset)
	if err != nil {
		return nil, errors.Wrap(apiutil.ErrValidation, err)
	}
	req := listPatsReq{
		token:  apiutil.ExtractBearerToken(r),
		limit:  l,
		offset: o,
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}

func decodeDeleteRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	return deletePatReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}, nil
}

func decodeResetSecretRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := resetPatSecretReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}

func decodeRevokeSecretRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	return revokePatSecretReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}, nil
}

func decodeAddScopeEntryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := addPatScopeEntryReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}

func decodeRemoveScopeEntryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := removePatScopeEntryReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}

func decodeClearAllScopeEntryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	return clearAllScopeEntryReq{
		token: apiutil.ExtractBearerToken(r),
		id:    chi.URLParam(r, "id"),
	}, nil
}

func decodeTestCheckScopeEntryRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, apiutil.ErrUnsupportedContentType
	}

	req := testCheckPatScopeReq{token: apiutil.ExtractBearerToken(r)}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(errors.ErrMalformedEntity, err)
	}
	return req, nil
}
