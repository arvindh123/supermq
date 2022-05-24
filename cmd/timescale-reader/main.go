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
	initutil "github.com/mainflux/mainflux/internal/init"
	"github.com/mainflux/mainflux/internal/init/mfserver"
	"github.com/mainflux/mainflux/internal/init/mfserver/httpserver"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/timescale"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "timescaledb-reader"
	stopWaitTime = 5 * time.Second

	defLogLevel          = "error"
	defPort              = "8911"
	defClientTLS         = "false"
	defCACerts           = ""
	defDBHost            = "localhost"
	defDBPort            = "5432"
	defDBUser            = "mainflux"
	defDBPass            = "mainflux"
	defDB                = "mainflux"
	defDBSSLMode         = "disable"
	defDBSSLCert         = ""
	defDBSSLKey          = ""
	defDBSSLRootCert     = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"

	envLogLevel          = "MF_TIMESCALE_READER_LOG_LEVEL"
	envPort              = "MF_TIMESCALE_READER_PORT"
	envClientTLS         = "MF_TIMESCALE_READER_CLIENT_TLS"
	envCACerts           = "MF_TIMESCALE_READER_CA_CERTS"
	envDBHost            = "MF_TIMESCALE_READER_DB_HOST"
	envDBPort            = "MF_TIMESCALE_READER_DB_PORT"
	envDBUser            = "MF_TIMESCALE_READER_DB_USER"
	envDBPass            = "MF_TIMESCALE_READER_DB_PASS"
	envDB                = "MF_TIMESCALE_READER_DB"
	envDBSSLMode         = "MF_TIMESCALE_READER_DB_SSL_MODE"
	envDBSSLCert         = "MF_TIMESCALE_READER_DB_SSL_CERT"
	envDBSSLKey          = "MF_TIMESCALE_READER_DB_SSL_KEY"
	envDBSSLRootCert     = "MF_TIMESCALE_READER_DB_SSL_ROOT_CERT"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel          string
	port              string
	clientTLS         bool
	caCerts           string
	dbConfig          timescale.Config
	jaegerURL         string
	thingsAuthURL     string
	usersAuthURL      string
	thingsAuthTimeout time.Duration
	usersAuthTimeout  time.Duration
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

	thingsTracer, thingsCloser := initutil.Jaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	authTracer, authCloser := initutil.Jaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := initutil.ConnectToAuth(cfg.clientTLS, cfg.caCerts, cfg.usersAuthURL, svcName, logger)
	defer authConn.Close()
	auth := authapi.NewClient(authTracer, authConn, cfg.usersAuthTimeout)

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	repo := newService(db, logger)

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(repo, tc, auth, svcName, logger), cfg.dbConfig.SSLCert, cfg.dbConfig.SSLKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return mfserver.ServerStopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Timescale reader service terminated: %s", err))
	}
}

func loadConfig() config {
	dbConfig := timescale.Config{
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

	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		port:              mainflux.Env(envPort, defPort),
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		dbConfig:          dbConfig,
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: authTimeout,
	}
}

func connectToDB(dbConfig timescale.Config, logger logger.Logger) *sqlx.DB {
	db, err := timescale.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Timescale: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, logger logger.Logger) readers.MessageRepository {
	svc := timescale.New(db)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := initutil.MakeMetrics("timescale", "message_reader")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
