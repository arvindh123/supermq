package mfserver

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mainflux/mainflux/logger"
)

type Server interface {
	Start() error
	Stop() error
}

type BaseServer struct {
	Ctx      context.Context
	Cancel   context.CancelFunc
	Name     string
	Address  string
	Port     string
	CertFile string
	KeyFile  string
	Logger   logger.Logger
	Protocol string
}

func stopAllServer(servers ...Server) error {
	var err error
	for _, server := range servers {
		err1 := server.Stop()
		if err1 != nil {
			if err == nil {
				err = fmt.Errorf("%w", err1)
			} else {
				err = fmt.Errorf("%v ; %w", err, err1)
			}
		}
	}
	return err
}

func ServerStopSignalHandler(ctx context.Context, cancel context.CancelFunc, logger logger.Logger, servers ...Server) error {
	var err error
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGABRT)
	select {
	case sig := <-c:
		defer cancel()
		err = stopAllServer(servers...)
		if err != nil {
			logger.Error(fmt.Sprintf("Authentication service error during shutdown: %v", err))
		}
		logger.Info(fmt.Sprintf("Authentication service shutdown by signal: %s", sig))
		return err
	case <-ctx.Done():
		return nil
	}
}
