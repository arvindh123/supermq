// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/influxdb"
	"github.com/mainflux/mainflux/internal"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"golang.org/x/sync/errgroup"
)

const (
	svcName = "influxdb-writer"

	defNatsURL    = "nats://localhost:4222"
	defLogLevel   = "error"
	defPort       = "8180"
	defDB         = "mainflux"
	defDBHost     = "localhost"
	defDBPort     = "8086"
	defDBUser     = "mainflux"
	defDBPass     = "mainflux"
	defConfigPath = "/config.toml"

	envNatsURL    = "MF_NATS_URL"
	envLogLevel   = "MF_INFLUX_WRITER_LOG_LEVEL"
	envPort       = "MF_INFLUX_WRITER_PORT"
	envDB         = "MF_INFLUXDB_DB"
	envDBHost     = "MF_INFLUXDB_HOST"
	envDBPort     = "MF_INFLUXDB_PORT"
	envDBUser     = "MF_INFLUXDB_ADMIN_USER"
	envDBPass     = "MF_INFLUXDB_ADMIN_PASSWORD"
	envConfigPath = "MF_INFLUX_WRITER_CONFIG_PATH"
)

type config struct {
	natsURL    string
	logLevel   string
	port       string
	dbName     string
	dbHost     string
	dbPort     string
	dbUser     string
	dbPass     string
	configPath string
}

func main() {
	cfg, clientCfg := loadConfigs()
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

	client, err := influxdata.NewHTTPClient(clientCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	defer client.Close()

	repo := newService(client, cfg.dbName, logger)

	if err := consumers.Start(svcName, pubSub, repo, cfg.configPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start InfluxDB writer: %s", err))
		os.Exit(1)
	}

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(svcName), "", "", logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("InfluxDB reader service terminated: %s", err))
	}
}

func loadConfigs() (config, influxdata.HTTPConfig) {
	cfg := config{
		natsURL:    mainflux.Env(envNatsURL, defNatsURL),
		logLevel:   mainflux.Env(envLogLevel, defLogLevel),
		port:       mainflux.Env(envPort, defPort),
		dbName:     mainflux.Env(envDB, defDB),
		dbHost:     mainflux.Env(envDBHost, defDBHost),
		dbPort:     mainflux.Env(envDBPort, defDBPort),
		dbUser:     mainflux.Env(envDBUser, defDBUser),
		dbPass:     mainflux.Env(envDBPass, defDBPass),
		configPath: mainflux.Env(envConfigPath, defConfigPath),
	}

	clientCfg := influxdata.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.dbHost, cfg.dbPort),
		Username: cfg.dbUser,
		Password: cfg.dbPass,
	}

	return cfg, clientCfg
}

func newService(client influxdata.Client, dbName string, logger logger.Logger) consumers.Consumer {
	repo := influxdb.New(client, dbName)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("influxdb", "message_writer")
	repo = api.MetricsMiddleware(repo, counter, latency)

	return repo
}
