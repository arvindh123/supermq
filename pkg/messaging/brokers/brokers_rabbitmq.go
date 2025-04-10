// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

//go:build rabbitmq
// +build rabbitmq

package brokers

import (
	"context"
	"log"
	"log/slog"

	"github.com/absmach/supermq/pkg/messaging"
	"github.com/absmach/supermq/pkg/messaging/rabbitmq"
)

func init() {
	log.Println("The binary was build using RabbitMQ as the message broker")
}

func NewPublisher(_ context.Context, typ messaging.PubSubType, url string, opts ...messaging.Option) (messaging.Publisher, error) {
	pb, err := rabbitmq.NewPublisher(typ, url, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func NewPubSub(_ context.Context, typ messaging.PubSubType, url string, logger *slog.Logger, opts ...messaging.Option) (messaging.PubSub, error) {
	pb, err := rabbitmq.NewPubSub(typ, url, logger, opts...)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

// AllSubjects represents subject to subscribe for all the Rabbit subject of given pubsub type.
func AllSubjects(typ messaging.PubSubType) (string, error) {
	conf, err := typ.Conf()
	if err != nil {
		return "", err
	}
	return conf.Rabbit.AllSubjects, nil
}
