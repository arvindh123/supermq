// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package apiutil

import (
	"fmt"
	"strings"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

func MakeMetrics(svcName string) (*kitprometheus.Counter, *kitprometheus.Summary) {
	var namespace, subsystem string
	parts := strings.Split(svcName, "-")
	namespace = parts[0]
	if len(parts) == 2 {
		// Reader
		if parts[1] == "reader" {
			counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: namespace,
				Subsystem: fmt.Sprintf("message_%s", parts[1]),
				Name:      "request_count",
				Help:      "Number of requests received.",
			}, []string{"method"})
			latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: namespace,
				Subsystem: fmt.Sprintf("message_%s", parts[1]),
				Name:      "request_latency_microseconds",
				Help:      "Total duration of requests in microseconds.",
			}, []string{"method"})
			return counter, latency
		}
		// Writer
		counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: fmt.Sprintf("message_%s", parts[1]),
			Name:      "request_count",
			Help:      "Number of database inserts.",
		}, []string{"method"})
		latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: namespace,
			Subsystem: fmt.Sprintf("message_%s", parts[1]),
			Name:      "request_latency_microseconds",
			Help:      "Total duration of inserts in microseconds.",
		}, []string{"method"})
		return counter, latency
	}
	subsystem = "api"
	namespace = svcName
	if svcName == "smtp-notifier" {
		subsystem = "smtp"
		namespace = "notifier"
	} else if svcName == "smpp-notifier" {
		subsystem = "smpp"
		namespace = "notifier"
	}
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_count",
		Help:      "Number of requests received.",
	}, []string{"method"})
	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "request_latency_microseconds",
		Help:      "Total duration of requests in microseconds.",
	}, []string{"method"})

	return counter, latency
}
