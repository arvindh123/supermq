package coap

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/mainflux/internal/server"
	"github.com/mainflux/mainflux/logger"
	gocoap "github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/mux"
)

const (
	stopWaitTime = 5 * time.Second
)

type Server struct {
	server.BaseServer
	handler mux.HandlerFunc
}

var _ server.Server = (*Server)(nil)

func New(ctx context.Context, cancel context.CancelFunc, name string, address string, port string, handler mux.HandlerFunc, logger logger.Logger) server.Server {
	return &Server{
		BaseServer: server.BaseServer{
			Ctx:     ctx,
			Cancel:  cancel,
			Name:    name,
			Address: address,
			Port:    fmt.Sprintf(":%s", port),
			Logger:  logger,
		},
		handler: handler,
	}
}

func (s *Server) Start() error {
	errCh := make(chan error)
	s.Logger.Info(fmt.Sprintf("%s service started using http, exposed port %s", s.Name, s.Port))
	go func() {
		errCh <- gocoap.ListenAndServe("udp", s.Port, s.handler)
	}()

	select {
	case <-s.Ctx.Done():
		return s.Stop()
	case err := <-errCh:
		return err
	}

}

func (s *Server) Stop() error {
	defer s.Cancel()
	c := make(chan bool)
	defer close(c)
	select {
	case <-c:
	case <-time.After(stopWaitTime):
	}
	s.Logger.Info(fmt.Sprintf("%s service shutdown of http at %s", s.Name, s.Port))
	return nil
}
