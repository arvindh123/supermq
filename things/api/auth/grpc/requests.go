// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import "github.com/mainflux/mainflux/internal"

type accessByKeyReq struct {
	thingKey string
	chanID   string
}

func (req accessByKeyReq) validate() error {
	if req.chanID == "" {
		return internal.ErrMissingID
	}

	if req.thingKey == "" {
		return internal.ErrBearerKey
	}

	return nil
}

type accessByIDReq struct {
	thingID string
	chanID  string
}

func (req accessByIDReq) validate() error {
	if req.thingID == "" || req.chanID == "" {
		return internal.ErrMissingID
	}

	return nil
}

type channelOwnerReq struct {
	owner  string
	chanID string
}

func (req channelOwnerReq) validate() error {
	if req.owner == "" || req.chanID == "" {
		return internal.ErrMissingID
	}

	return nil
}

type identifyReq struct {
	key string
}

func (req identifyReq) validate() error {
	if req.key == "" {
		return internal.ErrBearerKey
	}

	return nil
}
