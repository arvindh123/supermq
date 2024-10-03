// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/auth"
	"github.com/absmach/magistrala/pkg/clients"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/absmach/magistrala/pkg/policies"
	"github.com/absmach/magistrala/things"
)

var _ things.Service = (*authorizationMiddleware)(nil)

type authorizationMiddleware struct {
	svc        things.Service
	authClient auth.AuthClient
}

// AuthorizationMiddleware adds authorization to the clients service.
func AuthorizationMiddleware(svc things.Service, authClient auth.AuthClient) things.Service {
	return &authorizationMiddleware{
		svc:        svc,
		authClient: authClient,
	}
}

func (am *authorizationMiddleware) CreateThings(ctx context.Context, session auth.Session, client ...clients.Client) ([]clients.Client, error) {
	if err := am.authorize(ctx, "", policies.UserType, policies.UsersKind, session.DomainUserID, policies.CreatePermission, policies.DomainType, session.DomainID); err != nil {
		return nil, err
	}

	return am.svc.CreateThings(ctx, session, client...)
}

func (am *authorizationMiddleware) ViewClient(ctx context.Context, session auth.Session, id string) (clients.Client, error) {
	if session.DomainUserID == "" {
		return clients.Client{}, svcerr.ErrDomainAuthorization
	}
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.ViewPermission, policies.ThingType, id); err != nil {
		return clients.Client{}, err
	}

	return am.svc.ViewClient(ctx, session, id)
}

func (am *authorizationMiddleware) ViewClientPerms(ctx context.Context, session auth.Session, id string) ([]string, error) {
	return am.svc.ViewClientPerms(ctx, session, id)
}

func (am *authorizationMiddleware) ListClients(ctx context.Context, session auth.Session, reqUserID string, pm clients.Page) (clients.ClientsPage, error) {
	if session.DomainUserID == "" {
		return clients.ClientsPage{}, svcerr.ErrDomainAuthorization
	}
	switch {
	case reqUserID != "" && reqUserID != session.UserID:
		if err := am.authorize(ctx, "", policies.UserType, policies.UsersKind, session.DomainUserID, policies.AdminPermission, policies.DomainType, session.DomainID); err != nil {
			return clients.ClientsPage{}, err
		}
	default:
		err := am.checkSuperAdmin(ctx, session.UserID)
		switch {
		case err == nil:
			session.SuperAdmin = true
		default:
			if err := am.authorize(ctx, "", policies.UserType, policies.UsersKind, session.DomainUserID, policies.MembershipPermission, policies.DomainType, session.DomainID); err != nil {
				return clients.ClientsPage{}, err
			}
		}
	}

	return am.svc.ListClients(ctx, session, reqUserID, pm)
}

func (am *authorizationMiddleware) ListClientsByGroup(ctx context.Context, session auth.Session, groupID string, pm clients.Page) (clients.MembersPage, error) {
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, pm.Permission, policies.GroupType, groupID); err != nil {
		return clients.MembersPage{}, err
	}

	return am.svc.ListClientsByGroup(ctx, session, groupID, pm)
}

func (am *authorizationMiddleware) UpdateClient(ctx context.Context, session auth.Session, client clients.Client) (clients.Client, error) {
	if session.DomainUserID == "" {
		return clients.Client{}, svcerr.ErrDomainAuthorization
	}
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.EditPermission, policies.ThingType, client.ID); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateClient(ctx, session, client)
}

func (am *authorizationMiddleware) UpdateClientTags(ctx context.Context, session auth.Session, client clients.Client) (clients.Client, error) {
	if session.DomainUserID == "" {
		return clients.Client{}, svcerr.ErrDomainAuthorization
	}
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.EditPermission, policies.ThingType, client.ID); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateClientTags(ctx, session, client)
}

func (am *authorizationMiddleware) UpdateClientSecret(ctx context.Context, session auth.Session, id, key string) (clients.Client, error) {
	if session.DomainUserID == "" {
		return clients.Client{}, svcerr.ErrDomainAuthorization
	}
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.EditPermission, policies.ThingType, id); err != nil {
		return clients.Client{}, err
	}

	return am.svc.UpdateClientSecret(ctx, session, id, key)
}

func (am *authorizationMiddleware) EnableClient(ctx context.Context, session auth.Session, id string) (clients.Client, error) {
	if session.DomainUserID == "" {
		return clients.Client{}, svcerr.ErrDomainAuthorization
	}
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.DeletePermission, policies.ThingType, id); err != nil {
		return clients.Client{}, err
	}

	return am.svc.EnableClient(ctx, session, id)
}

func (am *authorizationMiddleware) DisableClient(ctx context.Context, session auth.Session, id string) (clients.Client, error) {
	if session.DomainUserID == "" {
		return clients.Client{}, svcerr.ErrDomainAuthorization
	}
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.DeletePermission, policies.ThingType, id); err != nil {
		return clients.Client{}, err
	}

	return am.svc.DisableClient(ctx, session, id)
}

func (am *authorizationMiddleware) Share(ctx context.Context, session auth.Session, id string, relation string, userids ...string) error {
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.DeletePermission, policies.ThingType, id); err != nil {
		return err
	}

	return am.svc.Share(ctx, session, id, relation, userids...)
}

func (am *authorizationMiddleware) Unshare(ctx context.Context, session auth.Session, id string, relation string, userids ...string) error {
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.DeletePermission, policies.ThingType, id); err != nil {
		return err
	}

	return am.svc.Unshare(ctx, session, id, relation, userids...)
}

func (am *authorizationMiddleware) Identify(ctx context.Context, key string) (string, error) {
	return am.svc.Identify(ctx, key)
}

func (am *authorizationMiddleware) DeleteClient(ctx context.Context, session auth.Session, id string) error {
	if err := am.authorize(ctx, session.DomainID, policies.UserType, policies.UsersKind, session.DomainUserID, policies.DeletePermission, policies.ThingType, id); err != nil {
		return err
	}

	return am.svc.DeleteClient(ctx, session, id)
}

func (am *authorizationMiddleware) checkSuperAdmin(ctx context.Context, adminID string) error {
	if _, err := am.authClient.Authorize(ctx, &magistrala.AuthorizeReq{
		SubjectType: policies.UserType,
		Subject:     adminID,
		Permission:  policies.AdminPermission,
		ObjectType:  policies.PlatformType,
		Object:      policies.MagistralaObject,
	}); err != nil {
		return err
	}
	return nil
}

func (am *authorizationMiddleware) authorize(ctx context.Context, domain, subjType, subjKind, subj, perm, objType, obj string) error {
	req := &magistrala.AuthorizeReq{
		Domain:      domain,
		SubjectType: subjType,
		SubjectKind: subjKind,
		Subject:     subj,
		Permission:  perm,
		ObjectType:  objType,
		Object:      obj,
	}
	res, err := am.authClient.Authorize(ctx, req)
	if err != nil {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	if !res.GetAuthorized() {
		return errors.Wrap(svcerr.ErrAuthorization, err)
	}
	return nil
}