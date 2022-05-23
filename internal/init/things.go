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

func ConnectToThings(clientTLS bool, caCerts, thingsAuthURL, svcName string, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	if clientTLS {
		if caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(caCerts, "")
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to load certs: %s", err))
				os.Exit(1)
			}
			opts = append(opts, grpc.WithTransportCredentials(tpc))
		}
	} else {
		logger.Info("gRPC communication is not encrypted")
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(thingsAuthURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to %s service: %s", svcName, err))
		os.Exit(1)
	}
	logger.Info(fmt.Sprintf("Established gRPC connection to %s via gRPC: %s", svcName, thingsAuthURL))
	return conn
}
