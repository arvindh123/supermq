// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package keys

import (
	"time"

	"github.com/mainflux/mainflux/auth"
	initutil "github.com/mainflux/mainflux/internal/init"
)

type issueKeyReq struct {
	token    string
	Type     uint32        `json:"type,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

// It is not possible to issue Reset key using HTTP API.
func (req issueKeyReq) validate() error {
	if req.token == "" {
		return initutil.ErrBearerToken
	}

	if req.Type != auth.LoginKey &&
		req.Type != auth.RecoveryKey &&
		req.Type != auth.APIKey {
		return initutil.ErrInvalidAPIKey
	}

	return nil
}

type keyReq struct {
	token string
	id    string
}

func (req keyReq) validate() error {
	if req.token == "" {
		return initutil.ErrBearerToken
	}

	if req.id == "" {
		return initutil.ErrMissingID
	}
	return nil
}
