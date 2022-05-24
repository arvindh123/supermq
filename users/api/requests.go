// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/users"
)

const (
	maxLimitSize = 100
	maxEmailSize = 1024
)

type userReq struct {
	user users.User
}

func (req userReq) validate() error {
	return req.user.Validate()
}

type createUserReq struct {
	user  users.User
	token string
}

func (req createUserReq) validate() error {
	return req.user.Validate()
}

type viewUserReq struct {
	token  string
	userID string
}

func (req viewUserReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	return nil
}

type listUsersReq struct {
	token    string
	offset   uint64
	limit    uint64
	email    string
	metadata users.Metadata
}

func (req listUsersReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.limit > maxLimitSize || req.limit < 1 {
		return internal.ErrLimitSize
	}

	if len(req.email) > maxEmailSize {
		return internal.ErrEmailSize
	}

	return nil
}

type updateUserReq struct {
	token    string
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateUserReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	return nil
}

type passwResetReq struct {
	Email string `json:"email"`
	Host  string `json:"host"`
}

func (req passwResetReq) validate() error {
	if req.Email == "" {
		return internal.ErrMissingEmail
	}

	if req.Host == "" {
		return internal.ErrMissingHost
	}

	return nil
}

type resetTokenReq struct {
	Token    string `json:"token"`
	Password string `json:"password"`
	ConfPass string `json:"confirm_password"`
}

func (req resetTokenReq) validate() error {
	if req.Password == "" {
		return internal.ErrMissingPass
	}

	if req.ConfPass == "" {
		return internal.ErrMissingConfPass
	}

	if req.Token == "" {
		return internal.ErrBearerToken
	}

	if req.Password != req.ConfPass {
		return internal.ErrInvalidResetPass
	}

	return nil
}

type passwChangeReq struct {
	token       string
	Password    string `json:"password"`
	OldPassword string `json:"old_password"`
}

func (req passwChangeReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}
	if req.OldPassword == "" {
		return internal.ErrMissingPass
	}
	return nil
}

type listMemberGroupReq struct {
	token    string
	offset   uint64
	limit    uint64
	metadata users.Metadata
	groupID  string
}

func (req listMemberGroupReq) validate() error {
	if req.token == "" {
		return internal.ErrBearerToken
	}

	if req.groupID == "" {
		return internal.ErrMissingID
	}

	return nil
}
