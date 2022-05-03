// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"fmt"
	"os"

	"github.com/mainflux/mainflux/logger"
	"google.golang.org/grpc"
)

// Keto intializes Keto read and write service
func Keto(hostReadAddress, readPort, hostWriteAddress, writePort string, logger logger.Logger) (readerConnection, writerConnection *grpc.ClientConn) {
	readConn, err := grpc.Dial(fmt.Sprintf("%s:%s", hostReadAddress, readPort), grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to dial %s:%s for Keto Read Service: %s", hostReadAddress, readPort, err))
		os.Exit(1)
	}

	writeConn, err := grpc.Dial(fmt.Sprintf("%s:%s", hostWriteAddress, writePort), grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to dial %s:%s for Keto Write Service: %s", hostWriteAddress, writePort, err))
		os.Exit(1)
	}

	return readConn, writeConn
}
