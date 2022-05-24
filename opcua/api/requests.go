// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import initutil "github.com/mainflux/mainflux/internal/init"

type browseReq struct {
	ServerURI  string
	Namespace  string
	Identifier string
}

func (req *browseReq) validate() error {
	if req.ServerURI == "" {
		return initutil.ErrMissingID
	}

	return nil
}
