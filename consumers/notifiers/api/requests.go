// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import initutil "github.com/mainflux/mainflux/internal/init"

type createSubReq struct {
	token   string
	Topic   string `json:"topic,omitempty"`
	Contact string `json:"contact,omitempty"`
}

func (req createSubReq) validate() error {
	if req.token == "" {
		return initutil.ErrBearerToken
	}
	if req.Topic == "" {
		return initutil.ErrInvalidTopic
	}
	if req.Contact == "" {
		return initutil.ErrInvalidContact
	}
	return nil
}

type subReq struct {
	token string
	id    string
}

func (req subReq) validate() error {
	if req.token == "" {
		return initutil.ErrBearerToken
	}
	if req.id == "" {
		return initutil.ErrMissingID
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
		return initutil.ErrBearerToken
	}
	return nil
}
