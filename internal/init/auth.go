// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package init

import (
	"fmt"
	"os"

	"github.com/mainflux/mainflux/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func ConnectToAuth(clientTLS bool, caCerts, usersAuthURL, svcName string, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	logger.Info("Connecting to auth via gRPC")
	if clientTLS {
		if caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(caCerts, "")
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to create tls credentials: %s", err))
				os.Exit(1)
			}
			opts = append(opts, grpc.WithTransportCredentials(tpc))
		}
	} else {
		opts = append(opts, grpc.WithInsecure())
		logger.Info("gRPC communication is not encrypted")
	}

	conn, err := grpc.Dial(usersAuthURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to %s service: %s", svcName, err))
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Established gRPC connection to %s via gRPC: %s", svcName, usersAuthURL))
	return conn
}
