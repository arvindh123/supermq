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

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/coap"
	"github.com/mainflux/mainflux/coap/api"
	"github.com/mainflux/mainflux/internal/apiutil"
	"github.com/mainflux/mainflux/internal/apiutil/mfserver"
	"github.com/mainflux/mainflux/internal/apiutil/mfserver/httpserver"
	logger "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	gocoap "github.com/plgd-dev/go-coap/v2"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	stopWaitTime = 5 * time.Second
	svcName      = "coap"

	defPort              = "5683"
	defNatsURL           = "nats://localhost:4222"
	defLogLevel          = "error"
	defClientTLS         = "false"
	defCACerts           = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"

	envPort              = "MF_COAP_ADAPTER_PORT"
	envNatsURL           = "MF_NATS_URL"
	envLogLevel          = "MF_COAP_ADAPTER_LOG_LEVEL"
	envClientTLS         = "MF_COAP_ADAPTER_CLIENT_TLS"
	envCACerts           = "MF_COAP_ADAPTER_CA_CERTS"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	port              string
	natsURL           string
	logLevel          string
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

	conn := connectToThings(cfg, logger)
	defer conn.Close()

	thingsTracer, thingsCloser := apiutil.InitJaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	nps, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nps.Close()

	svc := coap.New(tc, nps)

	svc = api.LoggingMiddleware(svc, logger)

	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "coap_adapter",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "coap_adapter",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	g.Go(func() error {
		return startCOAPServer(ctx, cfg, svc, nil, logger)
	})

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHTTPHandler(), "", "", logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return mfserver.ServerStopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("CoAP adapter service terminated: %s", err))
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
		port:              mainflux.Env(envPort, defPort),
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: authTimeout,
	}
}

func connectToThings(cfg config, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	if cfg.clientTLS {
		if cfg.caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(cfg.caCerts, "")
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to load certs: %s", err))
				os.Exit(1)
			}
			opts = append(opts, grpc.WithTransportCredentials(tpc))
		}
	} else {
		logger.Info("gRPC communication is not encrypted")
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(cfg.thingsAuthURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to things service: %s", err))
		os.Exit(1)
	}
	return conn
}

func startCOAPServer(ctx context.Context, cfg config, svc coap.Service, auth mainflux.ThingsServiceClient, l logger.Logger) error {
	p := fmt.Sprintf(":%s", cfg.port)
	errCh := make(chan error)
	l.Info(fmt.Sprintf("CoAP adapter service started, exposed port %s", cfg.port))
	go func() {
		errCh <- gocoap.ListenAndServe("udp", p, api.MakeCoAPHandler(svc, l))
	}()
	select {
	case <-ctx.Done():
		l.Info(fmt.Sprintf("CoAP adapter service shutdown of http at %s", p))
		return nil
	case err := <-errCh:
		return err
	}
}
