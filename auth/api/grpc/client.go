// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"
	"fmt"
	"time"

	"github.com/absmach/magistrala"
	"github.com/absmach/magistrala/pkg/errors"
	svcerr "github.com/absmach/magistrala/pkg/errors/service"
	"github.com/go-kit/kit/endpoint"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const domainsSvcName = "magistrala.DomainsService"

var _ magistrala.DomainsServiceClient = (*domainsGrpcClient)(nil)

type domainsGrpcClient struct {
	deleteUserFromDomains endpoint.Endpoint
	timeout               time.Duration
}

// NewDomainsClient returns new policy gRPC client instance.
func NewDomainsClient(conn *grpc.ClientConn, timeout time.Duration) magistrala.DomainsServiceClient {
	return &domainsGrpcClient{
		deleteUserFromDomains: kitgrpc.NewClient(
			conn,
			domainsSvcName,
			"DeleteUserFromDomains",
			encodeDeleteUserRequest,
			decodeDeleteUserResponse,
			magistrala.DeleteUserRes{},
		).Endpoint(),

		timeout: timeout,
	}
}

func (client domainsGrpcClient) DeleteUserFromDomains(ctx context.Context, in *magistrala.DeleteUserReq, opts ...grpc.CallOption) (*magistrala.DeleteUserRes, error) {
	ctx, cancel := context.WithTimeout(ctx, client.timeout)
	defer cancel()

	res, err := client.deleteUserFromDomains(ctx, deleteUserPoliciesReq{
		ID: in.GetId(),
	})
	if err != nil {
		return &magistrala.DeleteUserRes{}, decodeError(err)
	}

	dpr := res.(deleteUserRes)
	return &magistrala.DeleteUserRes{Deleted: dpr.deleted}, nil
}

func decodeDeleteUserResponse(_ context.Context, grpcRes interface{}) (interface{}, error) {
	res := grpcRes.(*magistrala.DeleteUserRes)
	return deleteUserRes{deleted: res.GetDeleted()}, nil
}

func encodeDeleteUserRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(deleteUserPoliciesReq)
	return &magistrala.DeleteUserReq{
		Id: req.ID,
	}, nil
}

func decodeError(err error) error {
	if st, ok := status.FromError(err); ok {
		switch st.Code() {
		case codes.NotFound:
			return errors.Wrap(svcerr.ErrNotFound, errors.New(st.Message()))
		case codes.InvalidArgument:
			return errors.Wrap(errors.ErrMalformedEntity, errors.New(st.Message()))
		case codes.AlreadyExists:
			return errors.Wrap(svcerr.ErrConflict, errors.New(st.Message()))
		case codes.Unauthenticated:
			return errors.Wrap(svcerr.ErrAuthentication, errors.New(st.Message()))
		case codes.OK:
			if msg := st.Message(); msg != "" {
				return errors.Wrap(errors.ErrUnidentified, errors.New(msg))
			}
			return nil
		case codes.FailedPrecondition:
			return errors.Wrap(errors.ErrMalformedEntity, errors.New(st.Message()))
		case codes.PermissionDenied:
			return errors.Wrap(svcerr.ErrAuthorization, errors.New(st.Message()))
		default:
			return errors.Wrap(fmt.Errorf("unexpected gRPC status: %s (status code:%v)", st.Code().String(), st.Code()), errors.New(st.Message()))
		}
	}
	return err
}
