// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package keys

import (
	"time"

	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/internal"
)

type issueKeyReq struct {
	token    string
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.Type != auth.LoginKey &&
		req.Type != auth.RecoveryKey &&
		req.Type != auth.APIKey {
		return internal.ErrInvalidAPIKey
	}

	return nil
}

type keyReq struct {
	token string
	id    string
}

func (req keyReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.id == "" {
		return internal.ErrMissingID
	}
	return nil
}
