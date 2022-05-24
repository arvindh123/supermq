package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	mfserver "github.com/mainflux/mainflux/internal/server"
	"github.com/mainflux/mainflux/logger"
)

const (
	stopWaitTime  = 5 * time.Second
	httpProtocol  = "http"
	httpsProtocol = "https"
)

type HTTPServer struct {
	mfserver.BaseServer
	server *http.Server
}

var _ mfserver.Server = (*HTTPServer)(nil)

func New(ctx context.Context, cancel context.CancelFunc, name string, address string, port string, handler http.Handler, certPath string, keyPath string, logger logger.Logger) mfserver.Server {
	listenFullAddress := fmt.Sprintf(":%s", port)
	server := &http.Server{Addr: listenFullAddress, Handler: handler}
	return &HTTPServer{
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
		server: server,
	}
}

func (s *HTTPServer) Start() error {
	errCh := make(chan error)
	s.Protocol = httpProtocol
	switch {
	case s.CertFile != "" || s.KeyFile != "":
		s.Logger.Info(fmt.Sprintf("%s service started using https, cert %s key %s, exposed port %s", s.Name, s.CertFile, s.KeyFile, s.Port))
		go func() {
			errCh <- s.server.ListenAndServeTLS(s.CertFile, s.KeyFile)
		}()
		s.Protocol = httpsProtocol
	default:
		s.Logger.Info(fmt.Sprintf("%s service started using http, exposed port %s", s.Name, s.Port))
		go func() {
			errCh <- s.server.ListenAndServe()
		}()
	}
	select {
	case <-s.Ctx.Done():
		return s.Stop()
	case err := <-errCh:
		return err
	}
}

func (s *HTTPServer) Stop() error {
	defer s.Cancel()
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), stopWaitTime)
	defer cancelShutdown()
	if err := s.server.Shutdown(ctxShutdown); err != nil {
		s.Logger.Error(fmt.Sprintf("%s %s service error occurred during shutdown at %s: %s", s.Name, s.Protocol, s.Port, err))
		return fmt.Errorf("%s %s service error occurred during shutdown at %s: %w", s.Name, s.Protocol, s.Port, err)
	}
	s.Logger.Info(fmt.Sprintf("%s %s service shutdown of http at %s", s.Name, s.Protocol, s.Port))
	return nil
}
