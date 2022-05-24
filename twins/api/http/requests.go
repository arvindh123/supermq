// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/twins"
)

const (
	maxNameSize  = 1024
	maxLimitSize = 100
)

type apiReq interface {
	validate() error
}

type addTwinReq struct {
	token      string
	Name       string                 `json:"name,omitempty"`
	Definition twins.Definition       `json:"definition,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (req addTwinReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if len(req.Name) > maxNameSize {
		return internal.ErrNameSize
	}

	return nil
}

type updateTwinReq struct {
	token      string
	id         string
	Name       string                 `json:"name,omitempty"`
	Definition twins.Definition       `json:"definition,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateTwinReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.id == "" {
		return internal.ErrMissingID
	}

	if len(req.Name) > maxNameSize {
		return internal.ErrNameSize
	}

	return nil
}

type viewTwinReq struct {
	token string
	id    string
}

func (req viewTwinReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.id == "" {
		return internal.ErrMissingID
	}

	return nil
}

type listReq struct {
	token    string
	offset   uint64
	limit    uint64
	name     string
	metadata map[string]interface{}
}

func (req *listReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.limit < 1 || req.limit > maxLimitSize {
		return internal.ErrLimitSize
	}

	if len(req.name) > maxNameSize {
		return internal.ErrNameSize
	}

	return nil
}

type listStatesReq struct {
	token  string
	offset uint64
	limit  uint64
	id     string
}

func (req *listStatesReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.id == "" {
		return internal.ErrMissingID
	}

	if req.limit == 0 || req.limit > maxLimitSize {
		return internal.ErrLimitSize
	}

	return nil
}
