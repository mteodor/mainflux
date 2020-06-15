// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/authn/api/grpc"
	"github.com/mainflux/mainflux/bootstrap"
	rediscons "github.com/mainflux/mainflux/bootstrap/redis/consumer"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/logger"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jmoiron/sqlx"
	"github.com/mainflux/mainflux/certs/api"
	"github.com/mainflux/mainflux/certs/postgres"
	mflog "github.com/mainflux/mainflux/logger"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

type Config struct {
	logLevel       string
	dbConfig       postgres.Config
	clientTLS      bool
	encKey         []byte
	caCerts        string
	httpPort       string
	serverCert     string
	serverKey      string
	baseURL        string
	thingsPrefix   string
	esThingsURL    string
	esThingsPass   string
	esThingsDB     string
	esURL          string
	esPass         string
	esDB           string
	esConsumerName string
	jaegerURL      string
	authnURL       string
	authnTimeout   time.Duration
}

func main() {
	cfg := loadConfig()

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	esClient := connectToRedis(cfg.esURL, cfg.esPass, cfg.esDB, logger)
	defer esClient.Close()

	authTracer, authCloser := initJaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := connectToAuth(cfg, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.authnTimeout)

	svc := newService(auth, db, logger, esClient, cfg)
	errs := make(chan error, 2)

	go startHTTPServer(svc, cfg, logger, errs)
	//go subscribeToThingsES(svc, thingsESConn, cfg.esConsumerName, logger)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Certs service terminated: %s", err))
}

func loadConfig() Config {
	return Config{}
}

func connectToRedis(redisURL, redisPass, redisDB string, logger mflog.Logger) *redis.Client {
	db, err := strconv.Atoi(redisDB)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to redis: %s", err))
		os.Exit(1)
	}

	return redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: redisPass,
		DB:       db,
	})
}
func connectToDB(dbConfig postgres.Config, logger logger.Logger) *sqlx.DB {
	db, err := postgres.Connect(dbConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to postgres: %s", err))
		os.Exit(1)
	}
	return db
}

func connectToAuth(cfg Config, logger logger.Logger) *grpc.ClientConn {
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

	conn, err := grpc.Dial(cfg.authnURL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to authn service: %s", err))
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

func newService(auth mainflux.AuthNServiceClient, db *sqlx.DB, logger mflog.Logger, esClient *redis.Client, cfg Config) certs.Service {
	certsRepo := postgres.NewCertsRepository(db, logger)

	config := mfsdk.Config{
		BaseURL:      cfg.baseURL,
		ThingsPrefix: cfg.thingsPrefix,
	}

	sdk := mfsdk.NewSDK(config)

	svc := certs.New(auth, certsRepo, sdk)
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

func startHTTPServer(svc certs.Service, cfg Config, logger mflog.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	if cfg.serverCert != "" || cfg.serverKey != "" {
		logger.Info(fmt.Sprintf("Bootstrap service started using https on port %s with cert %s key %s",
			cfg.httpPort, cfg.serverCert, cfg.serverKey))
		errs <- http.ListenAndServeTLS(p, cfg.serverCert, cfg.serverKey, api.MakeHandler(svc))
		return
	}
	logger.Info(fmt.Sprintf("Bootstrap service started using http on port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, api.MakeHandler(svc))
}

func subscribeToThingsES(svc bootstrap.Service, client *redis.Client, consumer string, logger mflog.Logger) {
	eventStore := rediscons.NewEventStore(svc, client, consumer, logger)
	logger.Info("Subscribed to Redis Event Store")
	if err := eventStore.Subscribe("mainflux.things"); err != nil {
		logger.Warn(fmt.Sprintf("Bootstrap service failed to subscribe to event sourcing: %s", err))
	}
}
