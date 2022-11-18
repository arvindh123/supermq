// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

//go:build !test

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/certs"
	log "github.com/mainflux/mainflux/logger"
)

var _ certs.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    certs.Service
}

// NewLoggingMiddleware adds logging facilities to the core service.
func NewLoggingMiddleware(svc certs.Service, logger log.Logger) certs.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) IssueCert(ctx context.Context, token, thingID, ttl string, keyBits int, keyType string) (c certs.Cert, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method issue_cert for token: %s and thing: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.IssueCert(ctx, token, thingID, ttl, keyBits, keyType)
}

func (lm *loggingMiddleware) ListCerts(ctx context.Context, token, thingID string, offset, limit uint64) (cp certs.Page, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_certs for token: %s and thing id: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListCerts(ctx, token, thingID, offset, limit)
}

func (lm *loggingMiddleware) ListSerials(ctx context.Context, token, thingID string, offset, limit uint64) (cp certs.Page, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method list_serials for token: %s and thing id: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ListSerials(ctx, token, thingID, offset, limit)
}

func (lm *loggingMiddleware) ViewCert(ctx context.Context, token, serialID string) (c certs.Cert, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method view_cert for token: %s and serial id %s took %s to complete", token, serialID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.ViewCert(ctx, token, serialID)
}

func (lm *loggingMiddleware) RevokeCert(ctx context.Context, token, thingID string) (c certs.Revoke, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method revoke_cert for token: %s and thing: %s took %s to complete", token, thingID, time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RevokeCert(ctx, token, thingID)
}

func (lm *loggingMiddleware) RenewCerts(ctx context.Context, bsUpdateRenewCert bool) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method renew_certs took %s to complete", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.RenewCerts(ctx, bsUpdateRenewCert)
}

func (lm *loggingMiddleware) AutoRenew(ctx context.Context, bsUpdateRenewCert bool, renewInterval time.Duration) (err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method auto_renew for complete , running for %s", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.AutoRenew(ctx, bsUpdateRenewCert, renewInterval)
}

func (lm *loggingMiddleware) AutoRevokeCerts(ctx context.Context, thingID string) (rev certs.Revoke, err error) {
	defer func(begin time.Time) {
		message := fmt.Sprintf("Method auto_renew for complete , running for %s", time.Since(begin))
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("%s with error: %s.", message, err))
			return
		}
		lm.logger.Info(fmt.Sprintf("%s without errors.", message))
	}(time.Now())

	return lm.svc.AutoRevokeCerts(ctx, thingID)
}
