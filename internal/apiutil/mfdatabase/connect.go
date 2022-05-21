// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mfdatabase

import (
	"context"
	"fmt"
	"os"
	"strconv"

	r "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToRedis(redisURL, redisPass, redisDB string, logger logger.Logger) *r.Client {
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to redis: %s", err))
		os.Exit(1)
	}

	return r.NewClient(&r.Options{
		Addr:     redisURL,
		Password: redisPass,
		DB:       db,
	})
}

func ConnectToMongoDB(host, port, name string, logger logger.Logger) *mongo.Database {
	addr := fmt.Sprintf("mongodb://%s:%s", host, port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(addr))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}

	return client.Database(name)
}
