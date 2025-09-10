package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5"
	"golang.org/x/sync/errgroup"

	grpcChannelsV1 "github.com/absmach/supermq/api/grpc/channels/v1"
	grpcClientsV1 "github.com/absmach/supermq/api/grpc/clients/v1"
	smqlog "github.com/absmach/supermq/logger"
	"github.com/absmach/supermq/pkg/authn"
	smqauthn "github.com/absmach/supermq/pkg/authn"
	authsvcAuthn "github.com/absmach/supermq/pkg/authn/authsvc"
	smqauthz "github.com/absmach/supermq/pkg/authz"
	authsvcAuthz "github.com/absmach/supermq/pkg/authz/authsvc"
	"github.com/absmach/supermq/pkg/connections"
	domainsAuthz "github.com/absmach/supermq/pkg/domains/grpcclient"
	"github.com/absmach/supermq/pkg/grpcclient"
	"github.com/absmach/supermq/pkg/server"
	httpserver "github.com/absmach/supermq/pkg/server/http"
)

const (
	defSvcHTTPPort    = "9007"
	svcName           = "rabbit-auth"
	envPrefixHTTP     = "SMQ_RABBIT_AUTH_HTTP_"
	envPrefixAuth     = "SMQ_AUTH_GRPC_"
	envPrefixChannels = "SMQ_CHANNELS_GRPC_"
	envPrefixClients  = "SMQ_CLIENTS_GRPC_"
	envPrefixDomains  = "SMQ_DOMAINS_GRPC_"
)

// Regex to parse routing_key: m.<domain_id>.c.<channel_id>.<subtopic>
var topicRegex = regexp.MustCompile(`^m\.([^\.]+)\.c\.([^\.]+)(?:\..*)?$`)

type config struct {
	LogLevel      string  `env:"SMQ_RABBIT_AUTH_LOG_LEVEL"                     envDefault:"info"`
	JaegerURL     url.URL `env:"SMQ_JAEGER_URL"                                envDefault:"http://localhost:4318/v1/traces"`
	BrokerURL     string  `env:"SMQ_MESSAGE_BROKER_URL"                        envDefault:"nats://localhost:4222"`
	SendTelemetry bool    `env:"SMQ_SEND_TELEMETRY"                            envDefault:"true"`
	InstanceID    string  `env:"SMQ_RABBIT_AUTH_INSTANCE_ID"                   envDefault:""`
	TraceRatio    float64 `env:"SMQ_JAEGER_TRACE_RATIO"                        envDefault:"1.0"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	g, ctx := errgroup.WithContext(ctx)

	defer cancel()

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load %s configuration : %s", svcName, err)
	}

	logger, err := smqlog.New(os.Stdout, cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to init logger: %s", err.Error())
	}

	var exitCode int
	defer smqlog.ExitWithError(&exitCode)

	// Initializing authn
	authGrpcCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&authGrpcCfg, env.Options{Prefix: envPrefixAuth}); err != nil {
		logger.Error(fmt.Sprintf("failed to load auth gRPC client configuration : %s", err))
		exitCode = 1
		return
	}
	authn, authnClient, err := authsvcAuthn.NewAuthentication(ctx, authGrpcCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authnClient.Close()
	logger.Info("AuthN  successfully connected to auth gRPC server " + authnClient.Secure())

	// Initializing domain authz
	domsGrpcCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&domsGrpcCfg, env.Options{Prefix: envPrefixDomains}); err != nil {
		logger.Error(fmt.Sprintf("failed to load domains gRPC client configuration : %s", err))
		exitCode = 1
		return
	}

	domAuthz, _, domainsHandler, err := domainsAuthz.NewAuthorization(ctx, domsGrpcCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer domainsHandler.Close()
	logger.Info("Domain AuthZ  successfully connected to domain gRPC server " + authnClient.Secure())

	// Initializing authz
	authz, authzClient, err := authsvcAuthz.NewAuthorization(ctx, authGrpcCfg, domAuthz)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer authzClient.Close()
	logger.Info("AuthZ  successfully connected to auth gRPC server " + authnClient.Secure())

	// Channel gRPC
	channelsClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&channelsClientCfg, env.Options{Prefix: envPrefixChannels}); err != nil {
		logger.Error(fmt.Sprintf("failed to load channels gRPC client configuration : %s", err))
		exitCode = 1
		return
	}

	channelsClient, channelsHandler, err := grpcclient.SetupChannelsClient(ctx, channelsClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer channelsHandler.Close()
	logger.Info("Channels service gRPC client successfully connected to channels gRPC server " + channelsHandler.Secure())

	// Client gRPC
	clientsClientCfg := grpcclient.Config{}
	if err := env.ParseWithOptions(&clientsClientCfg, env.Options{Prefix: envPrefixClients}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s auth configuration : %s", svcName, err))
		exitCode = 1
		return
	}

	clientsClient, clientsHandler, err := grpcclient.SetupClientsClient(ctx, clientsClientCfg)
	if err != nil {
		logger.Error(err.Error())
		exitCode = 1
		return
	}
	defer clientsHandler.Close()
	logger.Info("Clients service gRPC client successfully connected to clients gRPC server " + clientsHandler.Secure())

	// Create auth struct
	a := newAuth(authn, authz, clientsClient, channelsClient)

	// HTTP server
	httpServerConfig := server.Config{Port: defSvcHTTPPort}
	if err := env.ParseWithOptions(&httpServerConfig, env.Options{Prefix: envPrefixHTTP}); err != nil {
		logger.Error(fmt.Sprintf("failed to load %s HTTP server configuration : %s", svcName, err))
		exitCode = 1
		return
	}
	httpSvc := httpserver.NewServer(ctx, cancel, svcName, httpServerConfig, a, logger)

	// Start all servers
	g.Go(func() error {
		return httpSvc.Start()
	})

	g.Go(func() error {
		return server.StopSignalHandler(ctx, cancel, logger, svcName, httpSvc)
	})

	if err := g.Wait(); err != nil {
		logger.Error(fmt.Sprintf("%s service terminated: %s", svcName, err))
	}

}

type auth struct {
	router   chi.Router
	authn    smqauthn.Authentication
	authz    smqauthz.Authorization
	clients  grpcClientsV1.ClientsServiceClient
	channels grpcChannelsV1.ChannelsServiceClient
}

func newAuth(authn smqauthn.Authentication, authz smqauthz.Authorization, clients grpcClientsV1.ClientsServiceClient, channels grpcChannelsV1.ChannelsServiceClient) *auth {

	a := &auth{
		router:   chi.NewRouter(),
		authn:    authn,
		authz:    authz,
		clients:  clients,
		channels: channels,
	}

	// register routes
	a.router.Post("/auth/user", a.userHandler)
	a.router.Post("/auth/topic", a.topicHandler)
	a.router.Post("/auth/vhost", a.vhostHandler)
	a.router.Post("/auth/resource", a.resourceHandler)
	a.router.NotFound(a.notfound)

	return a
}

func (a *auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.router.ServeHTTP(w, r)
}

func (a auth) HTTPHandler() http.Handler {
	r := chi.NewRouter()
	r.Post("/auth/user", a.userHandler)
	r.Post("/auth/topic", a.topicHandler)
	r.Post("/auth/vhost", a.vhostHandler)
	r.Post("/auth/resource", a.resourceHandler)
	r.NotFound(a.notfound)
	return r

}

func (a auth) notfound(w http.ResponseWriter, r *http.Request) {
	// Log request method and path
	log.Printf("[NotFound] Method=%s Path=%s", r.Method, r.URL.Path)

	// Log query parameters
	if len(r.URL.Query()) > 0 {
		log.Printf("[NotFound] QueryParams: %v", r.URL.Query())
	}

	// Log form parameters
	if err := r.ParseForm(); err == nil && len(r.PostForm) > 0 {
		log.Printf("[NotFound] FormParams: %v", r.PostForm)
	}

	// Read and log body
	bodyBytes, _ := io.ReadAll(r.Body)
	if len(bodyBytes) > 0 {
		log.Printf("[NotFound] Body: %s", string(bodyBytes))
		// Restore body for future readers
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	// Respond with 404
	http.Error(w, "not found", http.StatusNotFound)
}

// userHandler handles RabbitMQ's user_path requests
func (a auth) userHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	log.Printf("[User - /auth/user] username=%s, password=%s", username, password)

	if a.checkUser(username, password) {
		w.Write([]byte("allow"))
	} else {
		w.Write([]byte("deny"))
	}
}

// topicHandler handles RabbitMQ's topic_path requests
func (a auth) topicHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	routingKey := r.FormValue("routing_key")
	permission := r.FormValue("permission") // "read" or "write"

	log.Printf("[Topic - /auth/topic] username=%s, routingKey=%s, permission=%s", username, routingKey, permission)

	var connType connections.ConnType
	switch permission {
	case "read":
		connType = connections.Subscribe
	case "write":
		connType = connections.Publish
	}

	domainID, channelID, ok := parseRoutingKey(routingKey)
	if !ok {
		log.Printf("Invalid routing key: %s", routingKey)
		w.Write([]byte("deny"))
		return
	}

	if a.checkTopic(username, "client", permission, domainID, channelID, routingKey, connType) {
		w.Write([]byte("allow"))
	} else {
		w.Write([]byte("deny"))
	}
}

// vhost authz handler.
func (a auth) vhostHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	vhost := r.FormValue("vhost")
	ip := r.FormValue("ip")

	log.Printf("[Vhost - /auth/vhost] user=%s vhost=%s ip=%s", username, vhost, ip)

	if a.checkVhost(username, vhost, ip) {
		w.Write([]byte("allow"))
	} else {
		w.Write([]byte("deny"))
	}
}

// Resource authz handler.
func (a auth) resourceHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	vhost := r.FormValue("vhost")
	resource := r.FormValue("resource")     // queue, exchange, topic
	name := r.FormValue("name")             // resource name
	permission := r.FormValue("permission") // configure, read, write

	log.Printf("[Resource - /auth/resource] user=%s vhost=%s res=%s name=%s perm=%s", username, vhost, resource, name, permission)
	if a.checkResource(username, vhost, resource, name, permission) {
		w.Write([]byte("allow"))
	} else {
		w.Write([]byte("deny"))
	}
}

// checkUser validates username/password (dummy check)
func (a auth) checkUser(username, password string) bool {
	resp, err := a.clients.Authenticate(context.Background(), &grpcClientsV1.AuthnReq{Token: authn.AuthPack(authn.BasicAuth, username, password)})
	if err != nil {
		log.Printf("[CheckUser] - username=%s password=%s error: %v", username, password, err)
		return false
	}
	if !resp.GetAuthenticated() {
		log.Printf("[CheckUser] - username=%s password=%s error: %s", username, password, "not authenticated")
		return false
	}

	return true
}

// checkTopic decides if user can publish/subscribe to given domain/channel
func (a auth) checkTopic(clientID, clientType, permission, domainID, channelID, routingKey string, connType connections.ConnType) bool {
	log.Printf("[CheckTopic - /auth/topic] clientID=%s clientType=%s Perm=%s Domain=%s Channel=%s Topic=%s connType=%s", clientID, clientType, permission, domainID, channelID, routingKey, connType)
	resp, err := a.channels.Authorize(context.Background(), &grpcChannelsV1.AuthzReq{
		DomainId:   domainID,
		ClientId:   clientID,
		ClientType: clientType,
		ChannelId:  channelID,
		Type:       uint32(connType),
	})

	if err != nil {
		log.Printf("[CheckTopic - /auth/topic] clientID=%s clientType=%s Perm=%s Domain=%s Channel=%s Topic=%s connType=%s error: %v", clientID, clientType, permission, domainID, channelID, routingKey, connType, err)
		return false
	}
	if !resp.GetAuthorized() {
		log.Printf("[CheckTopic - /auth/topic] clientID=%s clientType=%s Perm=%s Domain=%s Channel=%s Topic=%s connType=%s error: %s", clientID, clientType, permission, domainID, channelID, routingKey, connType, "not authorized")
		return false
	}
	return true
}

func (a auth) checkVhost(username, vhost, ip string) bool {
	return true
}

func (a auth) checkResource(username, vhost, resource, name, permission string) bool {
	return true
}

// parseRoutingKey extracts domain_id and channel_id
func parseRoutingKey(rk string) (domainID, channelID string, ok bool) {
	m := topicRegex.FindStringSubmatch(rk)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

func nameHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
