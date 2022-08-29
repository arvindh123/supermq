package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/auth"
	api "github.com/mainflux/mainflux/auth/api"
	grpcapi "github.com/mainflux/mainflux/auth/api/grpc"
	httpapi "github.com/mainflux/mainflux/auth/api/http"
	"github.com/mainflux/mainflux/auth/jwt"
	"github.com/mainflux/mainflux/auth/keto"
	"github.com/mainflux/mainflux/auth/postgres"
	"github.com/mainflux/mainflux/auth/tracing"
	"github.com/mainflux/mainflux/internal"
	internalauth "github.com/mainflux/mainflux/internal/auth"
	"github.com/mainflux/mainflux/internal/server"
	grpcserver "github.com/mainflux/mainflux/internal/server/grpc"
	httpserver "github.com/mainflux/mainflux/internal/server/http"
	"github.com/mainflux/mainflux/internal/starter"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/uuid"
	"github.com/opentracing/opentracing-go"
	acl "github.com/ory/keto/proto/ory/keto/acl/v1alpha1"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	svcName          = "auth"
	defListenAddress = ""
)

type config struct {
	LogLevel      string        `env:"MF_AUTH_LOG_LEVEL"             default:"debug"`
	HttpPort      string        `env:"MF_AUTH_HTTP_PORT"             default:"6180"`
	GrpcPort      string        `env:"MF_AUTH_GRPC_PORT"             default:"6181"`
	Secret        string        `env:"MF_AUTH_SECRET"                default:"auth"`
	ServerCert    string        `env:"MF_AUTH_SERVER_CERT"           default:""`
	ServerKey     string        `env:"MF_AUTH_SERVER_KEY"            default:""`
	JaegerURL     string        `env:"MF_JAEGER_URL"                 default:""`
	KetoReadHost  string        `env:"MF_KETO_READ_REMOTE_HOST"      default:"mainflux-keto"`
	KetoWriteHost string        `env:"MF_KETO_WRITE_REMOTE_HOST"     default:"mainflux-keto"`
	KetoWritePort string        `env:"MF_KETO_READ_REMOTE_PORT"      default:"4466"`
	KetoReadPort  string        `env:"MF_KETO_WRITE_REMOTE_PORT"     default:"4467"`
	LoginDuration time.Duration `env:"MF_AUTH_LOGIN_TOKEN_DURATION"  default:"10h"`
}

type DbConfig struct {
	Host        string `env:"MF_AUTH_DB_HOST"           default:"localhost"`
	Port        string `env:"MF_AUTH_DB_PORT"           default:"5432"`
	User        string `env:"MF_AUTH_DB_USER"           default:"mainflux"`
	Pass        string `env:"MF_AUTH_DB_PASS"           default:"mainflux"`
	Name        string `env:"MF_AUTH_DB"                default:"auth"`
	SSLMode     string `env:"MF_AUTH_DB_SSL_MODE"       default:"disable"`
	SSLCert     string `env:"MF_AUTH_DB_SSL_CERT"       default:""`
	SSLKey      string `env:"MF_AUTH_DB_SSL_KEY"        default:""`
	SSLRootCert string `env:"MF_AUTH_DB_SSL_ROOT_CERT"  default:""`
}

func main() {
	cfg := config{}
	dbConfig := DbConfig{}

	err := starter.LoadConfig(&cfg)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s configuration : %s", svcName, err.Error()))
	}
	err = starter.LoadConfig(&dbConfig)
	if err != nil {
		log.Fatalf(fmt.Sprintf("Failed to load %s database configuration : %s", svcName, err.Error()))
	}

	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)
	logger, err := logger.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(postgres.Config(dbConfig), logger)
	defer db.Close()

	tracer, closer := internalauth.Jaeger("auth", cfg.JaegerURL, logger)
	defer closer.Close()

	dbTracer, dbCloser := internalauth.Jaeger("auth_db", cfg.JaegerURL, logger)
	defer dbCloser.Close()

	readerConn, writerConn := internalauth.Keto(cfg.KetoReadHost, cfg.KetoReadPort, cfg.KetoWriteHost, cfg.KetoWritePort, logger)

	svc := newService(db, dbTracer, cfg.Secret, logger, readerConn, writerConn, cfg.LoginDuration)

	registerAuthServiceServer := func(srv *grpc.Server) {
		mainflux.RegisterAuthServiceServer(srv, grpcapi.NewServer(tracer, svc))
	}

	hs := httpserver.New(ctx, cancel, svcName, defListenAddress, cfg.HttpPort, httpapi.MakeHandler(svc, tracer, logger), cfg.ServerCert, cfg.ServerKey, logger)
	gs := grpcserver.New(ctx, cancel, svcName, defListenAddress, cfg.GrpcPort, registerAuthServiceServer, cfg.ServerCert, cfg.ServerKey, logger)
	g.Go(func() error {
		return hs.Start()
	})
	g.Go(func() error {
		return gs.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, hs, gs)
	})
	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("Authentication service terminated: %s", err))
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

func newService(db *sqlx.DB, tracer opentracing.Tracer, secret string, logger logger.Logger, readerConn, writerConn *grpc.ClientConn, duration time.Duration) auth.Service {
	database := postgres.NewDatabase(db)
	keysRepo := tracing.New(postgres.New(database), tracer)

	groupsRepo := postgres.NewGroupRepo(database)
	groupsRepo = tracing.GroupRepositoryMiddleware(tracer, groupsRepo)

	pa := keto.NewPolicyAgent(acl.NewCheckServiceClient(readerConn), acl.NewWriteServiceClient(writerConn), acl.NewReadServiceClient(readerConn))

	idProvider := uuid.New()
	t := jwt.New(secret)

	svc := auth.New(keysRepo, groupsRepo, idProvider, t, pa, duration)
	svc = api.LoggingMiddleware(svc, logger)

	counter, latency := internal.MakeMetrics(svcName, "api")
	svc = api.MetricsMiddleware(svc, counter, latency)

	return svc
}
