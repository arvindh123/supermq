// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/certs"
)

func issueCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addCertsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		res, err := svc.IssueCert(ctx, req.token, req.thingID, req.Name, req.TTL)
		if err != nil {
			return certsRes{}, err
		}

		return CertToCertResponse(res, true), nil
	}
}

func listSerials(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		page, err := svc.ListCerts(ctx, req.token, req.certID, req.thingID, req.serial, req.name, req.offset, req.limit)
		if err != nil {
			return certsPageRes{}, err
		}
		res := certsPageRes{
			pageRes: pageRes{
				Total:  page.Total,
				Offset: page.Offset,
				Limit:  page.Limit,
			},
			Certs: []certsRes{},
		}

		for _, cert := range page.Certs {
			cr := CertToCertResponse(cert, true)
			res.Certs = append(res.Certs, cr)
		}
		return res, nil
	}
}

func viewCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewRevokeRenewRemoveReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		cert, err := svc.ViewCert(ctx, req.token, req.certID)
		if err != nil {
			return certsPageRes{}, err
		}

		return CertToCertResponse(cert, false), nil
	}
}

func revokeCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewRevokeRenewRemoveReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return nil, svc.RevokeCert(ctx, req.token, req.certID)
	}
}

func renewCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewRevokeRenewRemoveReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return svc.RenewCert(ctx, req.token, req.certID)
	}
}

func removeCert(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(viewRevokeRenewRemoveReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return nil, svc.RemoveCert(ctx, req.token, req.certID)
	}
}

func revokeThingCerts(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokeRenewRemoveThingIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return nil, svc.RevokeThingCerts(ctx, req.token, req.thingID, req.limit)
	}
}

func renewThingCerts(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokeRenewRemoveThingIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return nil, svc.RenewThingCerts(ctx, req.token, req.thingID, req.limit)
	}
}

func removeThingCerts(svc certs.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokeRenewRemoveThingIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}
		return nil, svc.RemoveThingCerts(ctx, req.token, req.thingID, req.limit)
	}
}
