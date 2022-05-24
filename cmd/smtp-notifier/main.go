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

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/notifiers"
	"github.com/mainflux/mainflux/consumers/notifiers/api"
	"github.com/mainflux/mainflux/consumers/notifiers/postgres"
	"github.com/mainflux/mainflux/consumers/notifiers/smtp"
	"github.com/mainflux/mainflux/consumers/notifiers/tracing"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/email"
	"github.com/mainflux/mainflux/internal/server"
	mfhttpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/mainflux/pkg/ulid"
	opentracing "github.com/opentracing/opentracing-go"
	"golang.org/x/sync/errgroup"
)

const (
	svcName          = "smtp-notifier"
	stopWaitTime     = 5 * time.Second
	defLogLevel      = "error"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "mainflux"
	defDBPass        = "mainflux"
	defDB            = "subscriptions"
	defConfigPath    = "/config.toml"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""
	defHTTPPort      = "8906"
	defServerCert    = ""
	defServerKey     = ""
	defFrom          = ""
	defJaegerURL     = ""
	defNatsURL       = "nats://localhost:4222"

	defEmailHost        = "localhost"
	defEmailPort        = "25"
	defEmailUsername    = "root"
	defEmailPassword    = ""
	defEmailFromAddress = ""
	defEmailFromName    = ""
	defEmailTemplate    = "email.tmpl"

	defAuthTLS     = "false"
	defAuthCACerts = ""
	defAuthURL     = "localhost:8181"
	defAuthTimeout = "1s"

	envLogLevel      = "MF_SMTP_NOTIFIER_LOG_LEVEL"
	envDBHost        = "MF_SMTP_NOTIFIER_DB_HOST"
	envDBPort        = "MF_SMTP_NOTIFIER_DB_PORT"
	envDBUser        = "MF_SMTP_NOTIFIER_DB_USER"
	envDBPass        = "MF_SMTP_NOTIFIER_DB_PASS"
	envDB            = "MF_SMTP_NOTIFIER_DB"
	envConfigPath    = "MF_SMTP_NOTIFIER_CONFIG_PATH"
	envDBSSLMode     = "MF_SMTP_NOTIFIER_DB_SSL_MODE"
	envDBSSLCert     = "MF_SMTP_NOTIFIER_DB_SSL_CERT"
	envDBSSLKey      = "MF_SMTP_NOTIFIER_DB_SSL_KEY"
	envDBSSLRootCert = "MF_SMTP_NOTIFIER_DB_SSL_ROOT_CERT"
	envHTTPPort      = "MF_SMTP_NOTIFIER_PORT"
	envServerCert    = "MF_SMTP_NOTIFIER_SERVER_CERT"
	envServerKey     = "MF_SMTP_NOTIFIER_SERVER_KEY"
	envFrom          = "MF_SMTP_NOTIFIER_FROM_ADDR"
	envJaegerURL     = "MF_JAEGER_URL"
	envNatsURL       = "MF_NATS_URL"

	envEmailHost        = "MF_EMAIL_HOST"
	envEmailPort        = "MF_EMAIL_PORT"
	envEmailUsername    = "MF_EMAIL_USERNAME"
	envEmailPassword    = "MF_EMAIL_PASSWORD"
	envEmailFromAddress = "MF_EMAIL_FROM_ADDRESS"
	envEmailFromName    = "MF_EMAIL_FROM_NAME"
	envEmailTemplate    = "MF_SMTP_NOTIFIER_TEMPLATE"

	envAuthTLS     = "MF_AUTH_CLIENT_TLS"
	envAuthCACerts = "MF_AUTH_CA_CERTS"
	envAuthURL     = "MF_AUTH_GRPC_URL"
	envAuthTimeout = "MF_AUTH_GRPC_TIMEOUT"
)

type config struct {
	natsURL     string
	configPath  string
	logLevel    string
	dbConfig    postgres.Config
	emailConf   email.Config
	from        string
	httpPort    string
	serverCert  string
	serverKey   string
	jaegerURL   string
	authTLS     bool
	authCACerts string
	authURL     string
	authTimeout time.Duration
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	authTracer, closer := internal.Jaeger("auth", cfg.jaegerURL, logger)
	defer closer.Close()

	authConn := internal.ConnectToAuth(cfg.authTLS, cfg.authCACerts, cfg.authURL, svcName, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.authTimeout)

	tracer, closer := internal.Jaeger("smtp-notifier", cfg.jaegerURL, logger)
	defer closer.Close()

	dbTracer, dbCloser := internal.Jaeger("smtp-notifier_db", cfg.jaegerURL, logger)
	defer dbCloser.Close()

	svc := newService(db, dbTracer, auth, cfg, logger)

	if err = consumers.Start(svcName, pubSub, svc, cfg.configPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to create Postgres writer: %s", err))
	}

	hs := mfhttpserver.New(ctx, cancel, svcName, "", cfg.httpPort, api.MakeHandler(svc, tracer, logger), cfg.serverCert, cfg.serverKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.ServerStopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("SMTP notifier service terminated: %s", err))
	}

}

func loadConfig() config {
	authTimeout, err := time.ParseDuration(mainflux.Env(envAuthTimeout, defAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthTimeout, err.Error())
	}

	tls, err := strconv.ParseBool(mainflux.Env(envAuthTLS, defAuthTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envAuthTLS)
	}

	dbConfig := postgres.Config{
		Host:        mainflux.Env(envDBHost, defDBHost),
		Port:        mainflux.Env(envDBPort, defDBPort),
		User:        mainflux.Env(envDBUser, defDBUser),
		Pass:        mainflux.Env(envDBPass, defDBPass),
		Name:        mainflux.Env(envDB, defDB),
		SSLMode:     mainflux.Env(envDBSSLMode, defDBSSLMode),
		SSLCert:     mainflux.Env(envDBSSLCert, defDBSSLCert),
		SSLKey:      mainflux.Env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: mainflux.Env(envDBSSLRootCert, defDBSSLRootCert),
	}

	emailConf := email.Config{
		FromAddress: mainflux.Env(envEmailFromAddress, defEmailFromAddress),
		FromName:    mainflux.Env(envEmailFromName, defEmailFromName),
		Host:        mainflux.Env(envEmailHost, defEmailHost),
		Port:        mainflux.Env(envEmailPort, defEmailPort),
		Username:    mainflux.Env(envEmailUsername, defEmailUsername),
		Password:    mainflux.Env(envEmailPassword, defEmailPassword),
		Template:    mainflux.Env(envEmailTemplate, defEmailTemplate),
	}

	return config{
		logLevel:    mainflux.Env(envLogLevel, defLogLevel),
		natsURL:     mainflux.Env(envNatsURL, defNatsURL),
		configPath:  mainflux.Env(envConfigPath, defConfigPath),
		dbConfig:    dbConfig,
		emailConf:   emailConf,
		from:        mainflux.Env(envFrom, defFrom),
		httpPort:    mainflux.Env(envHTTPPort, defHTTPPort),
		serverCert:  mainflux.Env(envServerCert, defServerCert),
		serverKey:   mainflux.Env(envServerKey, defServerKey),
		jaegerURL:   mainflux.Env(envJaegerURL, defJaegerURL),
		authTLS:     tls,
		authCACerts: mainflux.Env(envAuthCACerts, defAuthCACerts),
		authURL:     mainflux.Env(envAuthURL, defAuthURL),
		authTimeout: authTimeout,
	}

}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, tracer opentracing.Tracer, auth mainflux.AuthServiceClient, c config, logger logger.Logger) notifiers.Service {
	database := postgres.NewDatabase(db)
	repo := tracing.New(postgres.New(database), tracer)
	idp := ulid.New()

	agent, err := email.New(&c.emailConf)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create email agent: %s", err))
		os.Exit(1)
	}

	notifier := smtp.New(agent)
	svc := notifiers.New(auth, repo, idp, notifier, c.from)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("notifier", "smtp")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
