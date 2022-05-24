// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	initutil "github.com/mainflux/mainflux/internal/init"
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
		return initutil.ErrBearerToken
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
		return initutil.ErrBearerToken
	}

	if req.limit > maxLimitSize || req.limit < 1 {
		return initutil.ErrLimitSize
	}

	if len(req.email) > maxEmailSize {
		return initutil.ErrEmailSize
	}

	return nil
}

type updateUserReq struct {
	token    string
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

func (req updateUserReq) validate() error {
	if req.token == "" {
		return initutil.ErrBearerToken
	}
	return nil
}

type passwResetReq struct {
	Email string `json:"email"`
	Host  string `json:"host"`
}

func (req passwResetReq) validate() error {
	if req.Email == "" {
		return initutil.ErrMissingEmail
	}

	if req.Host == "" {
		return initutil.ErrMissingHost
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
		return initutil.ErrMissingPass
	}

	if req.ConfPass == "" {
		return initutil.ErrMissingConfPass
	}

	if req.Token == "" {
		return initutil.ErrBearerToken
	}

	if req.Password != req.ConfPass {
		return initutil.ErrInvalidResetPass
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
		return initutil.ErrBearerToken
	}
	if req.OldPassword == "" {
		return initutil.ErrMissingPass
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
		return initutil.ErrBearerToken
	}

	if req.groupID == "" {
		return initutil.ErrMissingID
	}

	return nil
}
