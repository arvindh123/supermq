// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// Package http contains the domain concept definitions needed to support
// Mainflux http adapter service functionality.
package http

import (
	"context"

	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/things/policies"
)

// Service specifies coap service API.
type Service interface {
	// Publish Messssage
	Publish(ctx context.Context, token string, msg *messaging.Message) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	publisher messaging.Publisher
	things    policies.ThingsServiceClient
}

// New instantiates the HTTP adapter implementation.
func New(publisher messaging.Publisher, things policies.ThingsServiceClient) Service {
	return &adapterService{
		publisher: publisher,
		things:    things,
	}
}

func (as *adapterService) Publish(ctx context.Context, token string, msg *messaging.Message) error {
	ar := &policies.TAuthorizeReq{
		Sub:        token,
		Obj:        msg.Channel,
		Act:        policies.ReadAction,
		EntityType: policies.GroupEntityType,
	}
	thid, err := as.things.AuthorizeByKey(ctx, ar)
	if err != nil {
		return err
	}
	msg.Publisher = thid.GetValue()

	return as.publisher.Publish(ctx, msg.Channel, msg)
}
