// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build !rabbitmq
// +build !rabbitmq

package brokers

import (
	"context"
	"log"
	"log/slog"

	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/messaging/nats"
)

func init() {
	log.Println("The binary was build using Nats as the message broker")
}

func NewPublisher(ctx context.Context, typ messaging.PubSubType, url string, opts ...messaging.Option) (messaging.Publisher, error) {
	pb, err := nats.NewPublisher(ctx, typ, url, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewPubSub(ctx context.Context, typ messaging.PubSubType, url string, logger *slog.Logger, opts ...messaging.Option) (messaging.PubSub, error) {
	pb, err := nats.NewPubSub(ctx, typ, url, logger, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

// AllSubjects represents subject to subscribe for all the NATS subject of given pubsub type.
func AllSubjects(typ messaging.PubSubType) (string, error) {
	conf, err := typ.Conf()
	if err != nil {
		return "", err
	}
	return conf.Nats.AllSubjects, nil
}
