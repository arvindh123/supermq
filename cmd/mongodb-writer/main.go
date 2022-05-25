// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/consumers"
	"github.com/mainflux/mainflux/consumers/writers/api"
	"github.com/mainflux/mainflux/consumers/writers/mongodb"
	"github.com/mainflux/mainflux/internal"
	mfdatabase "github.com/mainflux/mainflux/internal/db"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "mongodb-writer"
	stopWaitTime = 5 * time.Second

	defLogLevel   = "error"
	defNatsURL    = "nats://localhost:4222"
	defPort       = "8180"
	defDB         = "mainflux"
	defDBHost     = "localhost"
	defDBPort     = "27017"
	defConfigPath = "/config.toml"

	envNatsURL    = "MF_NATS_URL"
	envLogLevel   = "MF_MONGO_WRITER_LOG_LEVEL"
	envPort       = "MF_MONGO_WRITER_PORT"
	envDB         = "MF_MONGO_WRITER_DB"
	envDBHost     = "MF_MONGO_WRITER_DB_HOST"
	envDBPort     = "MF_MONGO_WRITER_DB_PORT"
	envConfigPath = "MF_MONGO_WRITER_CONFIG_PATH"
)

type config struct {
	natsURL    string
	logLevel   string
	port       string
	dbName     string
	dbHost     string
	dbPort     string
	configPath string
}

func main() {
	cfg := loadConfigs()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatal(err)
	}

	pubSub, err := nats.NewPubSub(cfg.natsURL, "", logger)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer pubSub.Close()

	db := mfdatabase.ConnectToMongoDB(cfg.dbHost, cfg.dbPort, cfg.dbName, logger)

	repo := newService(db, logger)

	if err := consumers.Start(svcName, pubSub, repo, cfg.configPath, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start MongoDB writer: %s", err))
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
		logger.Error(fmt.Sprintf("MongoDB writer service terminated: %s", err))
	}

}

func loadConfigs() config {
	return config{
		natsURL:    mainflux.Env(envNatsURL, defNatsURL),
		logLevel:   mainflux.Env(envLogLevel, defLogLevel),
		port:       mainflux.Env(envPort, defPort),
		dbName:     mainflux.Env(envDB, defDB),
		dbHost:     mainflux.Env(envDBHost, defDBHost),
		dbPort:     mainflux.Env(envDBPort, defDBPort),
		configPath: mainflux.Env(envConfigPath, defConfigPath),
	}
}

func newService(db *mongo.Database, logger logger.Logger) consumers.Consumer {
	repo := mongodb.New(db)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("mongodb", "message_writer")
	repo = api.MetricsMiddleware(repo, counter, latency)

	return repo
}
