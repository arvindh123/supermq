// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	internaldb "github.com/mainflux/mainflux/internal/db"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	internaldb "github.com/mainflux/mainflux/internal/db"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/brokers"
	"github.com/mainflux/mainflux/pkg/uuid"
	localusers "github.com/mainflux/mainflux/things/standalone"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/api"
	twapi "github.com/mainflux/mainflux/twins/api/http"
	twmongodb "github.com/mainflux/mainflux/twins/mongodb"
	rediscache "github.com/mainflux/mainflux/twins/redis"
	"github.com/mainflux/mainflux/twins/tracing"
	opentracing "github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

const (
	svcName            = "twins"
	queue              = "twins"
	defLogLevel        = "error"
	defHTTPPort        = "8180"
	defJaegerURL       = ""
	defServerCert      = ""
	defServerKey       = ""
	defDB              = "mainflux-twins"
	defDBHost          = "localhost"
	defDBPort          = "27017"
	defCacheURL        = "localhost:6379"
	defCachePass       = ""
	defCacheDB         = "0"
	defStandaloneEmail = ""
	defStandaloneToken = ""
	defClientTLS       = "false"
	defCACerts         = ""
	defChannelID       = ""
	defBrokerURL       = "nats://localhost:4222"
	defAuthURL         = "localhost:8181"
	defAuthTimeout     = "1s"

	envLogLevel        = "MF_TWINS_LOG_LEVEL"
	envHTTPPort        = "MF_TWINS_HTTP_PORT"
	envJaegerURL       = "MF_JAEGER_URL"
	envServerCert      = "MF_TWINS_SERVER_CERT"
	envServerKey       = "MF_TWINS_SERVER_KEY"
	envDB              = "MF_TWINS_DB"
	envDBHost          = "MF_TWINS_DB_HOST"
	envDBPort          = "MF_TWINS_DB_PORT"
	envCacheURL        = "MF_TWINS_CACHE_URL"
	envCachePass       = "MF_TWINS_CACHE_PASS"
	envCacheDB         = "MF_TWINS_CACHE_DB"
	envStandaloneEmail = "MF_TWINS_STANDALONE_EMAIL"
	envStandaloneToken = "MF_TWINS_STANDALONE_TOKEN"
	envClientTLS       = "MF_TWINS_CLIENT_TLS"
	envCACerts         = "MF_TWINS_CA_CERTS"
	envChannelID       = "MF_TWINS_CHANNEL_ID"
	envBrokerURL       = "MF_BROKER_URL"
	envAuthURL         = "MF_AUTH_GRPC_URL"
	envAuthTimeout     = "MF_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel        string
	httpPort        string
	jaegerURL       string
	serverCert      string
	serverKey       string
	dbCfg           twmongodb.Config
	cacheURL        string
	cachePass       string
	cacheDB         string
	standaloneEmail string
	standaloneToken string
	clientTLS       bool
	caCerts         string
	channelID       string
	brokerURL       string

	authURL     string
	authTimeout time.Duration
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	cacheClient := internaldb.ConnectToRedis(cfg.cacheURL, cfg.cachePass, cfg.cacheDB, logger)
	cacheTracer, cacheCloser := internalauth.Jaeger("twins_cache", cfg.jaegerURL, logger)
	cacheClient := internaldb.ConnectToRedis(cfg.cacheURL, cfg.cachePass, cfg.cacheDB, logger)
	cacheTracer, cacheCloser := internalauth.Jaeger("twins_cache", cfg.jaegerURL, logger)
	defer cacheCloser.Close()

	db, err := twmongodb.Connect(cfg.dbCfg, logger)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	dbTracer, dbCloser := internalauth.Jaeger("twins_db", cfg.jaegerURL, logger)
	dbTracer, dbCloser := internalauth.Jaeger("twins_db", cfg.jaegerURL, logger)
	defer dbCloser.Close()

	authTracer, authCloser := internalauth.Jaeger("auth", cfg.jaegerURL, logger)
	authTracer, authCloser := internalauth.Jaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()
	auth, _ := createAuthClient(cfg, authTracer, logger)

	pubSub, err := brokers.NewPubSub(cfg.brokerURL, queue, logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to message broker: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	svc := newService(svcName, pubSub, cfg.channelID, auth, dbTracer, db, cacheTracer, cacheClient, logger)
	svc := newService(svcName, pubSub, cfg.channelID, auth, dbTracer, db, cacheTracer, cacheClient, logger)

	tracer, closer := internalauth.Jaeger("twins", cfg.jaegerURL, logger)
	tracer, closer := internalauth.Jaeger("twins", cfg.jaegerURL, logger)
	defer closer.Close()

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.httpPort, twapi.MakeHandler(tracer, svc, logger), cfg.serverCert, cfg.serverKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.httpPort, twapi.MakeHandler(tracer, svc, logger), cfg.serverCert, cfg.serverKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})
	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Twins service terminated: %s", err))
	}
}
	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Twins service terminated: %s", err))
	}
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envAuthTimeout, defAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthTimeout, err.Error())
	}

	dbCfg := twmongodb.Config{
		Name: mainflux.Env(envDB, defDB),
		Host: mainflux.Env(envDBHost, defDBHost),
		Port: mainflux.Env(envDBPort, defDBPort),
	}

	return config{
		logLevel:        mainflux.Env(envLogLevel, defLogLevel),
		httpPort:        mainflux.Env(envHTTPPort, defHTTPPort),
		serverCert:      mainflux.Env(envServerCert, defServerCert),
		serverKey:       mainflux.Env(envServerKey, defServerKey),
		jaegerURL:       mainflux.Env(envJaegerURL, defJaegerURL),
		dbCfg:           dbCfg,
		cacheURL:        mainflux.Env(envCacheURL, defCacheURL),
		cachePass:       mainflux.Env(envCachePass, defCachePass),
		cacheDB:         mainflux.Env(envCacheDB, defCacheDB),
		standaloneEmail: mainflux.Env(envStandaloneEmail, defStandaloneEmail),
		standaloneToken: mainflux.Env(envStandaloneToken, defStandaloneToken),
		clientTLS:       tls,
		caCerts:         mainflux.Env(envCACerts, defCACerts),
		channelID:       mainflux.Env(envChannelID, defChannelID),
		brokerURL:       mainflux.Env(envBrokerURL, defBrokerURL),
		authURL:         mainflux.Env(envAuthURL, defAuthURL),
		authTimeout:     authTimeout,
	}
}

func createAuthClient(cfg config, tracer opentracing.Tracer, logger logger.Logger) (mainflux.AuthServiceClient, func() error) {
	if cfg.standaloneEmail != "" && cfg.standaloneToken != "" {
		return localusers.NewAuthService(cfg.standaloneEmail, cfg.standaloneToken), nil
	}

	conn := internalauth.ConnectToAuth(cfg.clientTLS, cfg.caCerts, cfg.authURL, svcName, logger)
	conn := internalauth.ConnectToAuth(cfg.clientTLS, cfg.caCerts, cfg.authURL, svcName, logger)
	return authapi.NewClient(tracer, conn, cfg.authTimeout), conn.Close
}

func newService(id string, ps messaging.PubSub, chanID string, users mainflux.AuthServiceClient, dbTracer opentracing.Tracer, db *mongo.Database, cacheTracer opentracing.Tracer, cacheClient *redis.Client, logger logger.Logger) twins.Service {
func newService(id string, ps messaging.PubSub, chanID string, users mainflux.AuthServiceClient, dbTracer opentracing.Tracer, db *mongo.Database, cacheTracer opentracing.Tracer, cacheClient *redis.Client, logger logger.Logger) twins.Service {
	twinRepo := twmongodb.NewTwinRepository(db)
	twinRepo = tracing.TwinRepositoryMiddleware(dbTracer, twinRepo)

	stateRepo := twmongodb.NewStateRepository(db)
	stateRepo = tracing.StateRepositoryMiddleware(dbTracer, stateRepo)

	idProvider := uuid.New()
	twinCache := rediscache.NewTwinCache(cacheClient)
	twinCache = tracing.TwinCacheMiddleware(cacheTracer, twinCache)

	svc := twins.New(ps, users, twinRepo, twinCache, stateRepo, idProvider, chanID, logger)
	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "twins",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "twins",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)
	err := ps.Subscribe(id, brokers.SubjectAllChannels, handle(logger, chanID, svc))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	return svc
}

func handle(logger logger.Logger, chanID string, svc twins.Service) handlerFunc {
	return func(msg messaging.Message) error {
func handle(logger logger.Logger, chanID string, svc twins.Service) handlerFunc {
	return func(msg messaging.Message) error {
		if msg.Channel == chanID {
			return nil
		}

		if err := svc.SaveStates(&msg); err != nil {
			logger.Error(fmt.Sprintf("State save failed: %s", err))
			return err
		}

		return nil
	}
}

type handlerFunc func(msg messaging.Message) error

func (h handlerFunc) Handle(msg messaging.Message) error {
	return h(msg)
}

func (h handlerFunc) Cancel() error {
	return nil
type handlerFunc func(msg messaging.Message) error

func (h handlerFunc) Handle(msg messaging.Message) error {
	return h(msg)
}

func (h handlerFunc) Cancel() error {
	return nil
}
