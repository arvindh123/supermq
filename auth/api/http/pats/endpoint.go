// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package pats

import (
	"context"

	"github.com/absmach/magistrala/auth"
	"github.com/go-kit/kit/endpoint"
)

func createPatEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createPatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.Create(ctx, req.token, req.Name, req.Description, req.Duration, req.Scope)
		if err != nil {
			return nil, err
		}

		return createPatRes{pat}, nil
	}
}

func retrieveEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(retrievePatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.Retrieve(ctx, req.token, req.id)
		if err != nil {
			return nil, err
		}

		return retrievePatRes{pat}, nil
	}
}

func updateNameEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updatePatNameReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.UpdateName(ctx, req.token, req.id, req.Name)
		if err != nil {
			return nil, err
		}

		return updatePatNameRes{pat}, nil
	}
}

func updateDescriptionEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updatePatDescriptionReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.UpdateDescription(ctx, req.token, req.id, req.Description)
		if err != nil {
			return nil, err
		}

		return updatePatDescriptionRes{pat}, nil
	}
}

func listEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listPatsReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pm := auth.PATSPageMeta{
			Limit:  req.limit,
			Offset: req.offset,
		}
		patsPage, err := svc.List(ctx, req.token, pm)
		if err != nil {
			return nil, err
		}

		return listPatsRes{patsPage}, nil
	}
}

func deleteEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(deletePatReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.Delete(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return deletePatRes{}, nil
	}
}

func resetSecretEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(resetPatSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		pat, err := svc.ResetSecret(ctx, req.token, req.id, req.Duration)
		if err != nil {
			return nil, err
		}

		return resetPatSecretRes{pat}, nil
	}
}

func revokeSecretEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(revokePatSecretReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.RevokeSecret(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return revokePatSecretRes{}, nil
	}
}

func addScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(addPatScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		scope, err := svc.AddScopeEntry(ctx, req.token, req.id, req.PlatformEntityType, req.OptionalDomainID, req.OptionalDomainEntityType, req.Operation, req.EntityIDs...)
		if err != nil {
			return nil, err
		}

		return addPatScopeEntryRes{scope}, nil
	}
}

func removeScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(removePatScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		scope, err := svc.RemoveScopeEntry(ctx, req.token, req.id, req.PlatformEntityType, req.OptionalDomainID, req.OptionalDomainEntityType, req.Operation, req.EntityIDs...)
		if err != nil {
			return nil, err
		}
		return removePatScopeEntryRes{scope}, nil
	}

}

func clearAllScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(clearAllScopeEntryReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.ClearAllScopeEntry(ctx, req.token, req.id); err != nil {
			return nil, err
		}

		return clearAllScopeEntryRes{}, nil
	}
}

func testCheckScopeEntryEndpoint(svc auth.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(testCheckPatScopeReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		if err := svc.TestCheckScopeEntry(ctx, req.token, req.PlatformEntityType, req.OptionalDomainID, req.OptionalDomainEntityType, req.Operation, req.EntityIDs...); err != nil {
			return nil, err
		}

		return revokePatSecretRes{}, nil
	}
}
