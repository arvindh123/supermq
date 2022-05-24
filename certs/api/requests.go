// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import "github.com/mainflux/mainflux/internal"

const maxLimitSize = 100

type addCertsReq struct {
	token   string
	ThingID string `json:"thing_id"`
	KeyBits int    `json:"key_bits"`
	KeyType string `json:"key_type"`
	TTL     string `json:"ttl"`
}

func (req addCertsReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.ThingID == "" {
		return internal.ErrMissingID
	}

	if req.TTL == "" || req.KeyType == "" || req.KeyBits == 0 {
		return internal.ErrMissingCertData
	}

	return nil
}

type listReq struct {
	thingID string
	token   string
	offset  uint64
	limit   uint64
}

func (req *listReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	if req.limit > maxLimitSize {
		return internal.ErrLimitSize
	}
	return nil
}

type viewReq struct {
	serialID string
	token    string
}

func (req *viewReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	if req.serialID == "" {
		return internal.ErrMissingID
	}

	return nil
}

type revokeReq struct {
	token  string
	certID string
}

func (req *revokeReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.certID == "" {
		return internal.ErrMissingID
	}

	return nil
}
