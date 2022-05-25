// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"fmt"
	"os"
	"strconv"

	r "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux/logger"
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
