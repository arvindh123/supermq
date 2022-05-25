package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	"github.com/mainflux/mainflux/internal/server"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/influxdb"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"golang.org/x/sync/errgroup"
)

const (
	svcName      = "influxdb-reader"
	stopWaitTime = 5 * time.Second

	defLogLevel          = "error"
	defPort              = "8180"
	defDB                = "mainflux"
	defDBHost            = "localhost"
	defDBPort            = "8086"
	defDBUser            = "mainflux"
	defDBPass            = "mainflux"
	defClientTLS         = "false"
	defCACerts           = ""
	defServerCert        = ""
	defServerKey         = ""
	defJaegerURL         = ""
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"
	defUsersAuthURL      = "localhost:8181"
	defUsersAuthTimeout  = "1s"

	envLogLevel          = "MF_INFLUX_READER_LOG_LEVEL"
	envPort              = "MF_INFLUX_READER_PORT"
	envDB                = "MF_INFLUXDB_DB"
	envDBHost            = "MF_INFLUXDB_HOST"
	envDBPort            = "MF_INFLUXDB_PORT"
	envDBUser            = "MF_INFLUXDB_ADMIN_USER"
	envDBPass            = "MF_INFLUXDB_ADMIN_PASSWORD"
	envClientTLS         = "MF_INFLUX_READER_CLIENT_TLS"
	envCACerts           = "MF_INFLUX_READER_CA_CERTS"
	envServerCert        = "MF_INFLUX_READER_SERVER_CERT"
	envServerKey         = "MF_INFLUX_READER_SERVER_KEY"
	envJaegerURL         = "MF_JAEGER_URL"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
	envAuthURL           = "MF_AUTH_GRPC_URL"
	envUsersAuthTimeout  = "MF_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel          string
	port              string
	dbName            string
	dbHost            string
	dbPort            string
	dbUser            string
	dbPass            string
	clientTLS         bool
	caCerts           string
	serverCert        string
	serverKey         string
	jaegerURL         string
	thingsAuthURL     string
	usersAuthURL      string
	thingsAuthTimeout time.Duration
	usersAuthTimeout  time.Duration
}

func main() {
	cfg, clientCfg := loadConfigs()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}
	conn := internalauth.ConnectToThings(cfg.clientTLS, cfg.caCerts, cfg.thingsAuthURL, svcName, logger)
	defer conn.Close()

	thingsTracer, thingsCloser := internalauth.Jaeger("things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	authTracer, authCloser := internalauth.Jaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := internalauth.ConnectToAuth(cfg.clientTLS, cfg.caCerts, cfg.usersAuthURL, svcName, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.usersAuthTimeout)

	client, err := influxdata.NewHTTPClient(clientCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	defer client.Close()

	repo := newService(client, cfg.dbName, logger)

	hs := httpserver.New(ctx, cancel, svcName, "", cfg.port, api.MakeHandler(repo, tc, auth, svcName, logger), cfg.serverCert, cfg.serverKey, logger)
	g.Go(func() error {
		return hs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("InfluxDB reader service terminated: %s", err))
	}
}

func loadConfigs() (config, influxdata.HTTPConfig) {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	userAuthTimeout, err := time.ParseDuration(mainflux.Env(envUsersAuthTimeout, defUsersAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	cfg := config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		port:              mainflux.Env(envPort, defPort),
		dbName:            mainflux.Env(envDB, defDB),
		dbHost:            mainflux.Env(envDBHost, defDBHost),
		dbPort:            mainflux.Env(envDBPort, defDBPort),
		dbUser:            mainflux.Env(envDBUser, defDBUser),
		dbPass:            mainflux.Env(envDBPass, defDBPass),
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		serverCert:        mainflux.Env(envServerCert, defServerCert),
		serverKey:         mainflux.Env(envServerKey, defServerKey),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: authTimeout,
		usersAuthURL:      mainflux.Env(envAuthURL, defUsersAuthURL),
		usersAuthTimeout:  userAuthTimeout,
	}

	clientCfg := influxdata.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.dbHost, cfg.dbPort),
		Username: cfg.dbUser,
		Password: cfg.dbPass,
	}

	return cfg, clientCfg
}

func newService(client influxdata.Client, dbName string, logger logger.Logger) readers.MessageRepository {
	repo := influxdb.New(client, dbName)
	repo = api.LoggingMiddleware(repo, logger)
	counter, latency := internal.MakeMetrics("influxdb", "message_reader")
	repo = api.MetricsMiddleware(repo, counter, latency)

	return repo
}
