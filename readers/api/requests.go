// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	initutil "github.com/mainflux/mainflux/internal/init"
	"github.com/mainflux/mainflux/readers"
)

const maxLimitSize = 1000

type apiReq interface {
	validate() error
}

type listMessagesReq struct {
	chanID   string
	token    string
	key      string
	pageMeta readers.PageMetadata
}

func (req listMessagesReq) validate() error {
	if req.token == "" && req.key == "" {
		return initutil.ErrBearerToken
	}

	if req.chanID == "" {
		return initutil.ErrMissingID
	}

	if req.pageMeta.Limit < 1 || req.pageMeta.Limit > maxLimitSize {
		return initutil.ErrLimitSize
	}

	if req.pageMeta.Offset < 0 {
		return initutil.ErrOffsetSize
	}

	if req.pageMeta.Comparator != "" &&
		req.pageMeta.Comparator != readers.EqualKey &&
		req.pageMeta.Comparator != readers.LowerThanKey &&
		req.pageMeta.Comparator != readers.LowerThanEqualKey &&
		req.pageMeta.Comparator != readers.GreaterThanKey &&
		req.pageMeta.Comparator != readers.GreaterThanEqualKey {
		return initutil.ErrInvalidComparator
	}

	return nil
}
