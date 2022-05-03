package grpcserver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/internal/apiutil/mfserver"
	"github.com/mainflux/mainflux/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	stopWaitTime  = 5 * time.Second
	httpProtocol  = "http"
	httpsProtocol = "https"
)

type GRPCServer struct {
	mfserver.BaseServer
	server     *grpc.Server
	authSrvSvr mainflux.AuthServiceServer
}

type serviceRegister func(*grpc.Server, interface{})

var _ mfserver.Server = (*GRPCServer)(nil)

func New(ctx context.Context, cancel context.CancelFunc, name string, address string, port string, authSrvSvr mainflux.AuthServiceServer, certPath string, keyPath string, logger logger.Logger) mfserver.Server {
	return &GRPCServer{
		BaseServer: mfserver.BaseServer{
			Ctx:      ctx,
			Cancel:   cancel,
			Name:     name,
			Address:  address,
			Port:     port,
			CertFile: certPath,
			KeyFile:  keyPath,
			Logger:   logger,
		},
		authSrvSvr: authSrvSvr,
	}
}

func (s *GRPCServer) Start() error {
	p := fmt.Sprintf("%s:%s", s.Address, s.Port)
	errCh := make(chan error)

	listener, err := net.Listen("tcp", p)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", s.Port, err)
	}

	switch {
	case s.CertFile != "" || s.KeyFile != "":
		creds, err := credentials.NewServerTLSFromFile(s.CertFile, s.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load auth certificates: %w", err)
		}
		s.Logger.Info(fmt.Sprintf("Authentication gRPC service started using https on port %s with cert %s key %s", s.Port, s.CertFile, s.KeyFile))
		s.server = grpc.NewServer(grpc.Creds(creds))
	default:
		s.Logger.Info(fmt.Sprintf("Authentication gRPC service started using http on port %s", s.Port))
		s.server = grpc.NewServer()
	}

	mainflux.RegisterAuthServiceServer(s.server, s.authSrvSvr)

	s.Logger.Info(fmt.Sprintf("Authentication gRPC service started, exposed port %s", s.Port))
	go func() {
		errCh <- s.server.Serve(listener)
	}()

	select {
	case <-s.Ctx.Done():
		return nil
	case err := <-errCh:
		s.Cancel()
		return err
	}
}

func (s *GRPCServer) Stop() error {
	defer s.Cancel()
	c := make(chan bool)
	go func() {
		defer close(c)
		s.server.GracefulStop()
	}()
	select {
	case <-c:
	case <-time.After(stopWaitTime):
	}
	s.Logger.Info(fmt.Sprintf("Authentication gRPC service shutdown at %s", s.Port))

	return nil
}
