// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/mainflux/mainflux"
	adapter "github.com/mainflux/mainflux/http"
	"github.com/mainflux/mainflux/http/api"
	initutil "github.com/mainflux/mainflux/internal/init"
	"github.com/mainflux/mainflux/internal/init/mfserver"
	"github.com/mainflux/mainflux/internal/init/mfserver/httpserver"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "http_adapter"
	stopWaitTime = 5 * time.Second

	defLogLevel          = "error"
	defClientTLS         = "false"
	defCACerts           = ""
	defPort              = "8180"
	defNatsURL           = "nats://localhost:4222"
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"

	envLogLevel          = "MF_HTTP_ADAPTER_LOG_LEVEL"
	envClientTLS         = "MF_HTTP_ADAPTER_CLIENT_TLS"
	envCACerts           = "MF_HTTP_ADAPTER_CA_CERTS"
	envPort              = "MF_HTTP_ADAPTER_PORT"
	envNatsURL           = "MF_NATS_URL"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	natsURL           string
	logLevel          string
	port              string
	clientTLS         bool
	caCerts           string
	jaegerURL         string
	thingsAuthURL     string
	thingsAuthTimeout time.Duration
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	conn := initutil.ConnectToThings(cfg.clientTLS, cfg.caCerts, cfg.thingsAuthURL, svcName, logger)
	defer conn.Close()

	tracer, closer := initutil.Jaeger("http_adapter", cfg.jaegerURL, logger)
	defer closer.Close()

	thingsTracer, thingsCloser := initutil.Jaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	pub, err := nats.NewPublisher(cfg.natsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pub.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	svc := newService(pub, tc, logger)

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(svc, tracer, logger), "", "", logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return mfserver.ServerStopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("HTTP adapter service terminated: %s", err))
	}

}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		natsURL:           mainflux.Env(envNatsURL, defNatsURL),
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		port:              mainflux.Env(envPort, defPort),
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: authTimeout,
	}
}

func newService(pub nats.Publisher, tc mainflux.ThingsServiceClient, logger logger.Logger) adapter.Service {
	svc := adapter.New(pub, tc)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := initutil.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
