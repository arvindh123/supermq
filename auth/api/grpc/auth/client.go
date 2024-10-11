// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"time"

	"github.com/absmach/magistrala"
	grpcapi "github.com/absmach/magistrala/auth/api/grpc"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
)

const authSvcName = "magistrala.AuthService"

type authGrpcClient struct {
	authenticate endpoint.Endpoint
	authorize    endpoint.Endpoint
	timeout      time.Duration
}

var _ magistrala.AuthServiceClient = (*authGrpcClient)(nil)

// NewAuthClient returns new auth gRPC client instance.
func NewAuthClient(conn *grpc.ClientConn, timeout time.Duration) magistrala.AuthServiceClient {
	return &authGrpcClient{
		authenticate: kitgrpc.NewClient(
			conn,
			authSvcName,
			"Authenticate",
			encodeIdentifyRequest,
			decodeIdentifyResponse,
			magistrala.AuthenticateRes{},
		).Endpoint(),
		authorize: kitgrpc.NewClient(
			conn,
			authSvcName,
			"Authorize",
			encodeAuthorizeRequest,
			decodeAuthorizeResponse,
			magistrala.AuthorizeRes{},
		).Endpoint(),
		timeout: timeout,
	}
}

func (client authGrpcClient) Authenticate(ctx context.Context, token *magistrala.AuthenticateReq, _ ...grpc.CallOption) (*magistrala.AuthenticateRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.authenticate(ctx, authenticateReq{token: token.GetToken()})
	if err != nil {
		return &magistrala.AuthenticateRes{}, grpcapi.DecodeError(err)
	}
	ir := res.(authenticateRes)
	return &magistrala.AuthenticateRes{Id: ir.id, UserId: ir.userID, DomainId: ir.domainID}, nil
}

func encodeIdentifyRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(authenticateReq)
	return &magistrala.AuthenticateReq{Token: req.token}, nil
}

func decodeIdentifyResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.AuthenticateRes)
	return authenticateRes{id: res.GetId(), userID: res.GetUserId(), domainID: res.GetDomainId()}, nil
}

func (client authGrpcClient) Authorize(ctx context.Context, req *magistrala.AuthorizeReq, _ ...grpc.CallOption) (r *magistrala.AuthorizeRes, err error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.authorize(ctx, authReq{
		Domain:      req.GetDomain(),
		SubjectType: req.GetSubjectType(),
		Subject:     req.GetSubject(),
		SubjectKind: req.GetSubjectKind(),
		Relation:    req.GetRelation(),
		Permission:  req.GetPermission(),
		ObjectType:  req.GetObjectType(),
		Object:      req.GetObject(),
	})
	if err != nil {
		return &magistrala.AuthorizeRes{}, grpcapi.DecodeError(err)
	}

	ar := res.(authorizeRes)
	return &magistrala.AuthorizeRes{Authorized: ar.authorized, Id: ar.id}, nil
}

func decodeAuthorizeResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.AuthorizeRes)
	return authorizeRes{authorized: res.Authorized, id: res.Id}, nil
}

func encodeAuthorizeRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(authReq)
	return &magistrala.AuthorizeReq{
		Domain:      req.Domain,
		SubjectType: req.SubjectType,
		Subject:     req.Subject,
		SubjectKind: req.SubjectKind,
		Relation:    req.Relation,
		Permission:  req.Permission,
		ObjectType:  req.ObjectType,
		Object:      req.Object,
	}, nil
}
