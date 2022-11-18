// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-redis/redis/v8"
	r "github.com/go-redis/redis/v8"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	rediscons "github.com/mainflux/mainflux/certs/redis/consumer"

	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/certs/api"
	certsBSClient "github.com/mainflux/mainflux/certs/bootstrap"
	vault "github.com/mainflux/mainflux/certs/pki"
	"github.com/mainflux/mainflux/certs/postgres"
	"github.com/mainflux/mainflux/logger"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jmoiron/sqlx"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

const (
	stopWaitTime = 5 * time.Second

	defLogLevel              = "error"
	defDBHost                = "localhost"
	defDBPort                = "5432"
	defDBUser                = "mainflux"
	defDBPass                = "mainflux"
	defDB                    = "certs"
	defDBSSLMode             = "disable"
	defDBSSLCert             = ""
	defDBSSLKey              = ""
	defDBSSLRootCert         = ""
	defClientTLS             = "false"
	defCACerts               = ""
	defPort                  = "8204"
	defServerCert            = ""
	defServerKey             = ""
	defCertsURL              = "http://localhost"
	defThingsURL             = "http://things:8182"
	defJaegerURL             = ""
	defAuthURL               = "localhost:8181"
	defAuthTimeout           = "1s"
	defSignCAPath            = "ca.crt"
	defSignCAKeyPath         = "ca.key"
	defSignHoursValid        = "2048h"
	defSignRSABits           = "2048"
	defCertAutoRenew         = true
	defCertAutoRenewUpdateBS = true
	defStopSvcOnRenewErr     = true

	defBootstrapURL = "http://bootstrap:8202/things/configs/certs"
	defMFUser       = "test@email.com"
	defMFPass       = "123"

	defThingsESURL  = "localhost:6379"
	defThingsESPass = ""
	defThingsESDB   = "0"

	defESURL          = "localhost:6379"
	defESPass         = ""
	defESDB           = "0"
	defESConsumerName = "certs"

	defVaultHost       = ""
	defVaultRole       = "mainflux"
	defVaultToken      = ""
	defVaultPKIIntPath = "pki_int"

	envPort           = "MF_CERTS_HTTP_PORT"
	envLogLevel       = "MF_CERTS_LOG_LEVEL"
	envDBHost         = "MF_CERTS_DB_HOST"
	envDBPort         = "MF_CERTS_DB_PORT"
	envDBUser         = "MF_CERTS_DB_USER"
	envDBPass         = "MF_CERTS_DB_PASS"
	envDB             = "MF_CERTS_DB"
	envDBSSLMode      = "MF_CERTS_DB_SSL_MODE"
	envDBSSLCert      = "MF_CERTS_DB_SSL_CERT"
	envDBSSLKey       = "MF_CERTS_DB_SSL_KEY"
	envDBSSLRootCert  = "MF_CERTS_DB_SSL_ROOT_CERT"
	envClientTLS      = "MF_CERTS_CLIENT_TLS"
	envCACerts        = "MF_CERTS_CA_CERTS"
	envServerCert     = "MF_CERTS_SERVER_CERT"
	envServerKey      = "MF_CERTS_SERVER_KEY"
	envCertsURL       = "MF_SDK_CERTS_URL"
	envJaegerURL      = "MF_JAEGER_URL"
	envAuthURL        = "MF_AUTH_GRPC_URL"
	envAuthTimeout    = "MF_AUTH_GRPC_TIMEOUT"
	envThingsURL      = "MF_THINGS_URL"
	envSignCAPath     = "MF_CERTS_SIGN_CA_PATH"
	envSignCAKey      = "MF_CERTS_SIGN_CA_KEY_PATH"
	envSignHoursValid = "MF_CERTS_SIGN_HOURS_VALID"
	envSignRSABits    = "MF_CERTS_SIGN_RSA_BITS"

	envVaultHost       = "MF_CERTS_VAULT_HOST"
	envVaultPKIIntPath = "MF_VAULT_PKI_INT_PATH"
	envVaultRole       = "MF_VAULT_CA_ROLE_NAME"
	envVaultToken      = "MF_VAULT_TOKEN"

	envBootstrapURL = "MF_BOOTSTRAP_URL"
	envMFUser       = "MF_USER"
	envMFPass       = "MF_PASS"

	envThingsESURL  = "MF_THINGS_ES_URL"
	envThingsESPass = "MF_THINGS_ES_PASS"
	envThingsESDB   = "MF_THINGS_ES_DB"

	envESConsumerName = "MF_CERTS_EVENT_CONSUMER"

	envUsersToken = "MF_USERS_TOKEN"
)

var (
	errFailedCertLoading     = errors.New("failed to load certificate")
	errFailedCertDecode      = errors.New("failed to decode certificate")
	errCACertificateNotExist = errors.New("CA certificate does not exist")
	errCAKeyNotExist         = errors.New("CA certificate key does not exist")
)

type config struct {
	logLevel     string
	dbConfig     postgres.Config
	clientTLS    bool
	caCerts      string
	httpPort     string
	serverCert   string
	serverKey    string
	certsURL     string
	thingsURL    string
	jaegerURL    string
	authURL      string
	authTimeout  time.Duration
	bootstrapURL string
	mfUser       string
	mfPass       string

	esThingsURL    string
	esThingsPass   string
	esThingsDB     string
	esConsumerName string

	// Sign and issue certificates without 3rd party PKI
	signCAPath     string
	signCAKeyPath  string
	signRSABits    int
	signHoursValid string
	// 3rd party PKI API access settings
	pkiPath  string
	pkiToken string
	pkiHost  string
	pkiRole  string
}

func main() {
	cfg := loadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tlsCert, caCert, err := loadCertificates(cfg)
	if err != nil {
		logger.Error("Failed to load CA certificates for issuing client certs")
	}

	if cfg.pkiHost == "" {
		log.Fatalf("No host specified for PKI engine")
	}

	pkiClient, err := vault.NewVaultClient(cfg.pkiToken, cfg.pkiHost, cfg.pkiPath, cfg.pkiRole)
	if err != nil {
		log.Fatalf("Failed to configure client for PKI engine")
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	thingsESConn := connectToRedis(cfg.esThingsURL, cfg.esThingsPass, cfg.esThingsDB, logger)
	defer thingsESConn.Close()

	authTracer, authCloser := initJaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := connectToAuth(cfg, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.authTimeout)

	svc := newService(auth, db, logger, nil, tlsCert, caCert, cfg, pkiClient)

	go subscribeToThingsES(svc, thingsESConn, cfg.esConsumerName, logger)

	g.Go(func() error {
		return startHTTPServer(ctx, svc, cfg, logger)
	})

	g.Go(func() error {
		if sig := errors.SignalHandler(ctx); sig != nil {
			cancel()
			logger.Info(fmt.Sprintf("Certs service shutdown by signal: %s", sig))
		}
		return nil
	})

	g.Go(func() error {
		return svc.AutoRenew(ctx, true, 10*time.Minute)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Certs service terminated: %s", err))
	}
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		tls = false
	}
	dbConfig := postgres.Config{
		Host:        mainflux.Env(envDBHost, defDBHost),
		Port:        mainflux.Env(envDBPort, defDBPort),
		User:        mainflux.Env(envDBUser, defDBUser),
		Pass:        mainflux.Env(envDBPass, defDBPass),
		Name:        mainflux.Env(envDB, defDB),
		SSLMode:     mainflux.Env(envDBSSLMode, defDBSSLMode),
		SSLCert:     mainflux.Env(envDBSSLCert, defDBSSLCert),
		SSLKey:      mainflux.Env(envDBSSLKey, defDBSSLKey),
		SSLRootCert: mainflux.Env(envDBSSLRootCert, defDBSSLRootCert),
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envAuthTimeout, defAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthTimeout, err.Error())
	}

	signRSABits, err := strconv.Atoi(mainflux.Env(envSignRSABits, defSignRSABits))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envSignRSABits, err.Error())
	}

	return config{
		logLevel:    mainflux.Env(envLogLevel, defLogLevel),
		dbConfig:    dbConfig,
		clientTLS:   tls,
		caCerts:     mainflux.Env(envCACerts, defCACerts),
		httpPort:    mainflux.Env(envPort, defPort),
		serverCert:  mainflux.Env(envServerCert, defServerCert),
		serverKey:   mainflux.Env(envServerKey, defServerKey),
		certsURL:    mainflux.Env(envCertsURL, defCertsURL),
		thingsURL:   mainflux.Env(envThingsURL, defThingsURL),
		jaegerURL:   mainflux.Env(envJaegerURL, defJaegerURL),
		authURL:     mainflux.Env(envAuthURL, defAuthURL),
		authTimeout: authTimeout,

		esThingsURL:    mainflux.Env(envThingsESURL, defThingsESURL),
		esThingsPass:   mainflux.Env(envThingsESPass, defThingsESPass),
		esThingsDB:     mainflux.Env(envThingsESDB, defThingsESDB),
		esConsumerName: mainflux.Env(envESConsumerName, defESConsumerName),

		signCAKeyPath:  mainflux.Env(envSignCAKey, defSignCAKeyPath),
		signCAPath:     mainflux.Env(envSignCAPath, defSignCAPath),
		signHoursValid: mainflux.Env(envSignHoursValid, defSignHoursValid),
		signRSABits:    signRSABits,

		pkiToken: mainflux.Env(envVaultToken, defVaultToken),
		pkiPath:  mainflux.Env(envVaultPKIIntPath, defVaultPKIIntPath),
		pkiRole:  mainflux.Env(envVaultRole, defVaultRole),
		pkiHost:  mainflux.Env(envVaultHost, defVaultHost),
	}

}

func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func connectToAuth(cfg config, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	if cfg.clientTLS {
		if cfg.caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(cfg.caCerts, "")
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

	conn, err := grpc.Dial(cfg.authURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to auth service: %s", err))
		os.Exit(1)
	}

	return conn
}

func initJaeger(svcName, url string, logger logger.Logger) (opentracing.Tracer, io.Closer) {
	if url == "" {
		return opentracing.NoopTracer{}, ioutil.NopCloser(nil)
	}

	tracer, closer, err := jconfig.Configuration{
		ServiceName: svcName,
		Sampler: &jconfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jconfig.ReporterConfig{
			LocalAgentHostPort: url,
			LogSpans:           true,
		},
	}.NewTracer()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to init Jaeger client: %s", err))
		os.Exit(1)
	}

	return tracer, closer
}

func newService(auth mainflux.AuthServiceClient, db *sqlx.DB, logger mflog.Logger, esClient *redis.Client, tlsCert tls.Certificate, x509Cert *x509.Certificate, cfg config, pkiAgent vault.Agent) certs.Service {
	certsRepo := postgres.NewRepository(db, logger)

	certsConfig := certs.Config{
		LogLevel:       cfg.logLevel,
		ClientTLS:      cfg.clientTLS,
		CaCerts:        cfg.caCerts,
		HTTPPort:       cfg.httpPort,
		ServerCert:     cfg.serverCert,
		ServerKey:      cfg.serverKey,
		CertsURL:       cfg.certsURL,
		JaegerURL:      cfg.jaegerURL,
		AuthURL:        cfg.authURL,
		AuthTimeout:    cfg.authTimeout,
		SignTLSCert:    tlsCert,
		SignX509Cert:   x509Cert,
		SignHoursValid: cfg.signHoursValid,
		SignRSABits:    cfg.signRSABits,
		PKIToken:       cfg.pkiToken,
		PKIHost:        cfg.pkiHost,
		PKIPath:        cfg.pkiPath,
		PKIRole:        cfg.pkiRole,
	}

	config := mfsdk.Config{
		CertsURL:  cfg.certsURL,
		ThingsURL: cfg.thingsURL,
	}

	sdk := mfsdk.NewSDK(config)
	bsClient := certsBSClient.New(cfg.bootstrapURL, cfg.mfUser, cfg.mfPass, sdk)

	svc := certs.New(auth, certsRepo, sdk, bsClient, certsConfig, pkiAgent)
	svc = api.NewLoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "certs",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "certs",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)
	return svc
}

func startHTTPServer(ctx context.Context, svc certs.Service, cfg config, logger mflog.Logger) error {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	errCh := make(chan error)
	server := &http.Server{Addr: p, Handler: api.MakeHandler(svc, logger)}
	switch {
	case cfg.serverCert != "" || cfg.serverKey != "":
		logger.Info(fmt.Sprintf("Certs service started using https on port %s with cert %s key %s", cfg.httpPort, cfg.serverCert, cfg.serverKey))
		go func() {
			errCh <- server.ListenAndServeTLS(cfg.serverCert, cfg.serverKey)
		}()

	default:
		logger.Info(fmt.Sprintf("Certs service started using http on port %s", cfg.httpPort))
		go func() {
			errCh <- http.ListenAndServe(p, api.MakeHandler(svc, logger))
		}()
	}
	select {
	case <-ctx.Done():
		ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), stopWaitTime)
		defer cancelShutdown()
		if err := server.Shutdown(ctxShutdown); err != nil {
			logger.Error(fmt.Sprintf("Certs service error occurred during shutdown at %s: %s", p, err))
			return fmt.Errorf("certs service  error occurred during shutdown at %s: %w", p, err)
		}
		logger.Info(fmt.Sprintf("Certs service  shutdown of http at %s", p))
		return nil
	case err := <-errCh:
		return err
	}
}

func loadCertificates(conf config) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	if conf.signCAPath == "" || conf.signCAKeyPath == "" {
		return tlsCert, caCert, nil
	}

	if _, err := os.Stat(conf.signCAPath); os.IsNotExist(err) {
		return tlsCert, caCert, errCACertificateNotExist
	}

	if _, err := os.Stat(conf.signCAKeyPath); os.IsNotExist(err) {
		return tlsCert, caCert, errCAKeyNotExist
	}

	tlsCert, err := tls.LoadX509KeyPair(conf.signCAPath, conf.signCAKeyPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertLoading, err)
	}

	b, err := ioutil.ReadFile(conf.signCAPath)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertLoading, err)
	}

	block, _ := pem.Decode(b)
	if block == nil {
		log.Fatalf("No PEM data found, failed to decode CA")
	}

	caCert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return tlsCert, caCert, errors.Wrap(errFailedCertDecode, err)
	}

	return tlsCert, caCert, nil
}

func subscribeToThingsES(svc certs.Service, client *r.Client, consumer string, logger mflog.Logger) {
	eventStore := rediscons.NewEventStore(svc, client, consumer, logger)
	logger.Info("Subscribed to Redis Event Store")
	if err := eventStore.Subscribe(context.Background(), "mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("certs service failed to subscribe to event sourcing: %s", err))
	}
}

func connectToRedis(redisURL, redisPass, redisDB string, logger mflog.Logger) *r.Client {
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
