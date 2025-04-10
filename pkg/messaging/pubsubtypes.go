// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package messaging

import (
	"errors"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

var (
	ErrSelfPubSubType    = errors.New("type 'Self' must be configured explicitly and does not have a predefined configuration")
	ErrUnknownPubSubType = errors.New("unknown PubSubType")
)

type PubSubType uint8

type NatsConf struct {
	JsConf jetstream.StreamConfig
}

type RabbitConf struct {
	ExchangeName string
}
type PubSubConf struct {
	PubTopicPrefix string
	Rabbit         RabbitConf
	Nats           NatsConf
}

const (
	Self PubSubType = iota
	Msg
	Alarms
	Writer
)

var pubSubConfMap = map[PubSubType]PubSubConf{
	Msg: {

		PubTopicPrefix: "m",
		Rabbit: RabbitConf{
			ExchangeName: "messaging",
		},
		Nats: NatsConf{
			JsConf: jetstream.StreamConfig{
				Name:              "messaging",
				Description:       "SuperMQ stream for sending and receiving messaging",
				Subjects:          []string{"m.>"},
				Retention:         jetstream.LimitsPolicy,
				MaxMsgsPerSubject: 1e6,
				MaxAge:            time.Hour * 24,
				MaxMsgSize:        1024 * 1024,
				Discard:           jetstream.DiscardOld,
				Storage:           jetstream.FileStorage,
			},
		},
	},
	Alarms: {
		PubTopicPrefix: "a",
		Rabbit: RabbitConf{
			ExchangeName: "alarms",
		},
		Nats: NatsConf{
			JsConf: jetstream.StreamConfig{
				Name:              "alarms",
				Description:       "SuperMQ stream for sending and receiving alarms",
				Subjects:          []string{"a.>"},
				Retention:         jetstream.LimitsPolicy,
				MaxMsgsPerSubject: 1e6,
				MaxAge:            time.Hour * 24,
				MaxMsgSize:        1024 * 1024,
				Discard:           jetstream.DiscardOld,
				Storage:           jetstream.FileStorage,
			},
		},
	},
	Writer: {
		PubTopicPrefix: "w",
		Rabbit: RabbitConf{
			ExchangeName: "writer",
		},
		Nats: NatsConf{
			JsConf: jetstream.StreamConfig{
				Name:              "writer",
				Description:       "SuperMQ stream for sending to writer",
				Subjects:          []string{"w.>"},
				Retention:         jetstream.LimitsPolicy,
				MaxMsgsPerSubject: 1e6,
				MaxAge:            time.Hour * 24,
				MaxMsgSize:        1024 * 1024,
				Discard:           jetstream.DiscardOld,
				Storage:           jetstream.FileStorage,
			},
		},
	},
}

func (t PubSubType) Conf() (PubSubConf, error) {
	if t == Self {
		return PubSubConf{}, ErrSelfPubSubType
	}
	conf, ok := pubSubConfMap[t]
	if !ok {
		return PubSubConf{}, fmt.Errorf("%w : %d", ErrUnknownPubSubType, t)
	}
	return conf, nil
}
