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
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	mfdatabase "github.com/mainflux/mainflux/internal/db"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/mongodb"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "mongodb-reader"
	stopWaitTime = 5 * time.Second

	defLogLevel          = "error"
	defPort              = "8180"
	defDB                = "mainflux"
	defDBHost            = "localhost"
	defDBPort            = "27017"
	defClientTLS         = "false"
	defCACerts           = ""
	defServerCert        = ""
	defServerKey         = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"
	defUsersAuthURL      = "localhost:8181"
	defUsersAuthTimeout  = "1s"

	envLogLevel          = "MF_MONGO_READER_LOG_LEVEL"
	envPort              = "MF_MONGO_READER_PORT"
	envDB                = "MF_MONGO_READER_DB"
	envDBHost            = "MF_MONGO_READER_DB_HOST"
	envDBPort            = "MF_MONGO_READER_DB_PORT"
	envClientTLS         = "MF_MONGO_READER_CLIENT_TLS"
	envCACerts           = "MF_MONGO_READER_CA_CERTS"
	envServerCert        = "MF_MONGO_READER_SERVER_CERT"
	envServerKey         = "MF_MONGO_READER_SERVER_KEY"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
	envUsersAuthURL      = "MF_AUTH_GRPC_URL"
	envUsersAuthTimeout  = "MF_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel          string
	port              string
	dbName            string
	dbHost            string
	dbPort            string
	clientTLS         bool
	caCerts           string
	serverCert        string
	serverKey         string
	jaegerURL         string
	thingsAuthURL     string
	usersAuthURL      string
	thingsAuthTimeout time.Duration
	usersAuthTimeout  time.Duration
}

func main() {
	cfg := loadConfigs()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	conn := internalauth.ConnectToThings(cfg.clientTLS, cfg.caCerts, cfg.thingsAuthURL, svcName, logger)
	defer conn.Close()

	thingsTracer, thingsCloser := internalauth.Jaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	authTracer, authCloser := internalauth.Jaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := internalauth.ConnectToAuth(cfg.clientTLS, cfg.caCerts, cfg.usersAuthURL, svcName, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.usersAuthTimeout)

	db := mfdatabase.ConnectToMongoDB(cfg.dbHost, cfg.dbPort, cfg.dbName, logger)

	repo := newService(db, logger)

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(repo, tc, auth, svcName, logger), cfg.serverCert, cfg.serverKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("MongoDB reader service terminated: %s", err))
	}

}

func loadConfigs() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	usersAuthTimeout, err := time.ParseDuration(mainflux.Env(envUsersAuthTimeout, defUsersAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		port:              mainflux.Env(envPort, defPort),
		dbName:            mainflux.Env(envDB, defDB),
		dbHost:            mainflux.Env(envDBHost, defDBHost),
		dbPort:            mainflux.Env(envDBPort, defDBPort),
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		serverCert:        mainflux.Env(envServerCert, defServerCert),
		serverKey:         mainflux.Env(envServerKey, defServerKey),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		usersAuthURL:      mainflux.Env(envUsersAuthURL, defUsersAuthURL),
		thingsAuthTimeout: authTimeout,
		usersAuthTimeout:  usersAuthTimeout,
	}
}

func newService(db *mongo.Database, logger logger.Logger) readers.MessageRepository {
	repo := mongodb.New(db)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("mongodb", "message_reader")
	repo = api.MetricsMiddleware(repo, counter, latency)

	return repo
}
