package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/mainflux/mainflux/internal/server"
	"github.com/mainflux/mainflux/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	stopWaitTime = 5 * time.Second
)

type Server struct {
	server.BaseServer
	server          *grpc.Server
	registerService serviceRegister
}

type serviceRegister func(srv *grpc.Server)

var _ server.Server = (*Server)(nil)

func New(ctx context.Context, cancel context.CancelFunc, name string, address string, port string, registerService serviceRegister, certPath string, keyPath string, logger logger.Logger) server.Server {
	return &Server{
		BaseServer: server.BaseServer{
			Ctx:      ctx,
			Cancel:   cancel,
			Name:     name,
			Address:  address,
			Port:     port,
			CertFile: certPath,
			KeyFile:  keyPath,
			Logger:   logger,
		},
		registerService: registerService,
	}
}

func (s *Server) Start() error {
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
		s.Logger.Info(fmt.Sprintf("%s gRPC service started using https on port %s with cert %s key %s", s.Name, s.Port, s.CertFile, s.KeyFile))
		s.server = grpc.NewServer(grpc.Creds(creds))
	default:
		s.Logger.Info(fmt.Sprintf("%s gRPC service started using http on port %s", s.Name, s.Port))
		s.server = grpc.NewServer()
	}

	s.registerService(s.server)

	s.Logger.Info(fmt.Sprintf("%s gRPC service started, exposed port %s", s.Name, s.Port))
	go func() {
		errCh <- s.server.Serve(listener)
	}()

	select {
	case <-s.Ctx.Done():
		return s.Stop()
	case err := <-errCh:
		s.Cancel()
		return err
	}
}

func (s *Server) Stop() error {
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
	s.Logger.Info(fmt.Sprintf("%s gRPC service shutdown at %s", s.Name, s.Port))

	return nil
}