// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import "github.com/mainflux/mainflux/internal"

type createSubReq struct {
	token   string
	Topic   string `json:"topic,omitempty"`
	Contact string `json:"contact,omitempty"`
}

func (req createSubReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	if req.Topic == "" {
		return internal.ErrInvalidTopic
	}
	if req.Contact == "" {
		return internal.ErrInvalidContact
	}
	return nil
}

type subReq struct {
	token string
	id    string
}

func (req subReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	if req.id == "" {
		return internal.ErrMissingID
	}
	return nil
}

type listSubsReq struct {
	token   string
	topic   string
	contact string
	offset  uint
	limit   uint
}

func (req listSubsReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	return nil
}
