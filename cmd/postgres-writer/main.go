// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/postgres"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"golang.org/x/sync/errgroup"
)

const (
	svcName = "postgres-writer"
	sep     = ","

	defLogLevel      = "error"
	defNatsURL       = "nats://localhost:4222"
	defPort          = "8180"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "mainflux"
	defDBPass        = "mainflux"
	defDB            = "mainflux"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""
	defConfigPath    = "/config.toml"

	envNatsURL       = "MF_NATS_URL"
	envLogLevel      = "MF_POSTGRES_WRITER_LOG_LEVEL"
	envPort          = "MF_POSTGRES_WRITER_PORT"
	envDBHost        = "MF_POSTGRES_WRITER_DB_HOST"
	envDBPort        = "MF_POSTGRES_WRITER_DB_PORT"
	envDBUser        = "MF_POSTGRES_WRITER_DB_USER"
	envDBPass        = "MF_POSTGRES_WRITER_DB_PASS"
	envDB            = "MF_POSTGRES_WRITER_DB"
	envDBSSLMode     = "MF_POSTGRES_WRITER_DB_SSL_MODE"
	envDBSSLCert     = "MF_POSTGRES_WRITER_DB_SSL_CERT"
	envDBSSLKey      = "MF_POSTGRES_WRITER_DB_SSL_KEY"
	envDBSSLRootCert = "MF_POSTGRES_WRITER_DB_SSL_ROOT_CERT"
	envConfigPath    = "MF_POSTGRES_WRITER_CONFIG_PATH"
)

type config struct {
	natsURL    string
	logLevel   string
	port       string
	configPath string
	dbConfig   postgres.Config
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	repo := newService(db, logger)

	if err = consumers.Start(svcName, pubSub, repo, cfg.configPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to create Postgres writer: %s", err))
	}

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(svcName), "", "", logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Postgres writer service terminated: %s", err))
	}

}

func loadConfig() config {
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

	return config{
		natsURL:    mainflux.Env(envNatsURL, defNatsURL),
		logLevel:   mainflux.Env(envLogLevel, defLogLevel),
		port:       mainflux.Env(envPort, defPort),
		configPath: mainflux.Env(envConfigPath, defConfigPath),
		dbConfig:   dbConfig,
	}
}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to Postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func newService(db *sqlx.DB, logger logger.Logger) consumers.Consumer {
	svc := postgres.New(db)
	svc = api.LoggingMiddleware(svc, logger)
	counter, latency := internal.MakeMetrics("postgres", "message_writer")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
