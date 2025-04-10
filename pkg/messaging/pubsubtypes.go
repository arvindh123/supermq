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
	JsConf      jetstream.StreamConfig
	AllSubjects string
}

type RabbitConf struct {
	ExchangeName string
	AllSubjects  string
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
	Msg:    msgPubSubConf,
	Alarms: alarmsPubSubConf,
	Writer: writerPubSubConf,
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

var msgPubSubConf = PubSubConf{
	PubTopicPrefix: "m",
	Rabbit: RabbitConf{
		ExchangeName: "messaging",
		AllSubjects:  "m.#",
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
		AllSubjects: "m.>",
	},
}

var alarmsPubSubConf = PubSubConf{
	PubTopicPrefix: "a",
	Rabbit: RabbitConf{
		ExchangeName: "alarms",
		AllSubjects:  "a.#",
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
		AllSubjects: "a.>",
	},
}

var writerPubSubConf = PubSubConf{
	PubTopicPrefix: "w",
	Rabbit: RabbitConf{
		ExchangeName: "writer",
		AllSubjects:  "w.#",
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
		AllSubjects: "w.>",
	},
}
