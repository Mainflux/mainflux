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

	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/auth/api/grpc"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/rules"
	"github.com/mainflux/mainflux/rules/api"
	rulesapi "github.com/mainflux/mainflux/rules/api/http"
	thingsapi "github.com/mainflux/mainflux/things/api/auth/grpc"
	"github.com/mainflux/mainflux/things/users"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	opentracing "github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	jconfig "github.com/uber/jaeger-client-go/config"
)

const (
	defLogLevel          = "info"
	defHTTPPort          = "9099"
	defKuiperURL         = "http://localhost:9081"
	defKuiperPluginHost  = "tcp://localhost"
	defKuiperPluginPort  = "4222"
	defServerCert        = ""
	defServerKey         = ""
	defSingleUserEmail   = ""
	defSingleUserToken   = ""
	defClientTLS         = "false"
	defCACerts           = ""
	defJaegerURL         = ""
	defAuthURL           = "localhost:8181"
	defAuthTimeout       = "1s"
	defThingsAuthURL     = "localhost:8183"
	defThingsAuthTimeout = "1s"

	envLogLevel          = "MF_RULES_LOG_LEVEL"
	envHTTPPort          = "MF_RULES_HTTP_PORT"
	envServerCert        = "MF_RULES_SERVER_CERT"
	envServerKey         = "MF_RULES_SERVER_KEY"
	envSingleUserEmail   = "MF_RULES_SINGLE_USER_EMAIL"
	envSingleUserToken   = "MF_RULES_SINGLE_USER_TOKEN"
	envClientTLS         = "MF_RULES_CLIENT_TLS"
	envCACerts           = "MF_RULES_CA_CERTS"
	envKuiperURL         = "MF_KUIPER_URL"
	envKuiperPluginHost  = "MF_KUIPER_PLUGIN_HOST"
	envKuiperPluginPort  = "MF_KUIPER_PLUGIN_PORT"
	envJaegerURL         = "MF_JAEGER_URL"
	envAuthURL           = "MF_AUTH_GRPC_URL"
	envAuthTimeout       = "MF_AUTH_GRPC_TIMEOUT"
	envThingsAuthURL     = "MF_THINGS_AUTH_GRPC_URL"
	envThingsAuthTimeout = "MF_THINGS_AUTH_GRPC_TIMEOUT"
)

type config struct {
	logLevel         string
	httpPort         string
	kuiperURL        string
	kuiperPluginHost string
	kuiperPluginPort string
	jaegerURL        string
	serverCert       string
	serverKey        string
	singleUserEmail  string
	singleUserToken  string
	clientTLS        bool
	caCerts          string

	authURL           string
	authTimeout       time.Duration
	thingsAuthURL     string
	thingsAuthTimeout time.Duration
}

func main() {
	cfg := loadConfig()

	logger, err := logger.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	authTracer, authCloser := initJaeger("rules_auth", cfg.jaegerURL, logger)
	defer authCloser.Close()
	auth, close := createAuthClient(cfg, authTracer, logger)
	if close != nil {
		defer close()
	}

	tracer, reCloser := initJaeger("rules", cfg.jaegerURL, logger)
	defer reCloser.Close()

	conn := connectToGRPC(cfg.clientTLS, cfg.caCerts, cfg.thingsAuthURL, logger)
	defer conn.Close()

	thingsTracer, thingsCloser := initJaeger("rules_things", cfg.jaegerURL, logger)
	defer thingsCloser.Close()

	tc := thingsapi.NewClient(conn, thingsTracer, cfg.thingsAuthTimeout)

	kuiper := rules.NewKuiperSDK(cfg.kuiperURL, cfg.kuiperPluginHost, cfg.kuiperPluginPort)

	svc := newService(kuiper, auth, tc, logger)
	errs := make(chan error, 2)

	go startHTTPServer(rulesapi.MakeHandler(tracer, svc), cfg, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err = <-errs
	logger.Error(fmt.Sprintf("Rules engine service terminated: %s", err))
}

func loadConfig() config {
	tls, err := strconv.ParseBool(mainflux.Env(envClientTLS, defClientTLS))
	if err != nil {
		log.Fatalf("Invalid value passed for %s\n", envClientTLS)
	}

	authTimeout, err := time.ParseDuration(mainflux.Env(envAuthTimeout, defAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthTimeout, err.Error())
	}

	thAuthTimeout, err := time.ParseDuration(mainflux.Env(envThingsAuthTimeout, defThingsAuthTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envThingsAuthTimeout, err.Error())
	}

	return config{
		logLevel:          mainflux.Env(envLogLevel, defLogLevel),
		httpPort:          mainflux.Env(envHTTPPort, defHTTPPort),
		kuiperURL:         mainflux.Env(envKuiperURL, defKuiperURL),
		kuiperPluginHost:  mainflux.Env(envKuiperPluginHost, defKuiperPluginHost),
		kuiperPluginPort:  mainflux.Env(envKuiperPluginPort, defKuiperPluginPort),
		serverCert:        mainflux.Env(envServerCert, defServerCert),
		serverKey:         mainflux.Env(envServerKey, defServerKey),
		jaegerURL:         mainflux.Env(envJaegerURL, defJaegerURL),
		singleUserEmail:   mainflux.Env(envSingleUserEmail, defSingleUserEmail),
		singleUserToken:   mainflux.Env(envSingleUserToken, defSingleUserToken),
		authURL:           mainflux.Env(envAuthURL, defAuthURL),
		authTimeout:       authTimeout,
		clientTLS:         tls,
		caCerts:           mainflux.Env(envCACerts, defCACerts),
		thingsAuthURL:     mainflux.Env(envThingsAuthURL, defThingsAuthURL),
		thingsAuthTimeout: thAuthTimeout,
	}
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

func newService(kuiper rules.KuiperSDK, auth mainflux.AuthServiceClient, things mainflux.ThingsServiceClient, logger logger.Logger) rules.Service {
	svc := rules.New(kuiper, auth, things, logger)

	svc = api.LoggingMiddleware(svc, logger)
	svc = api.MetricsMiddleware(
		svc,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "rules",
			Subsystem: "api",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "rules",
			Subsystem: "api",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	return svc
}

func startHTTPServer(handler http.Handler, cfg config, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	if cfg.serverCert != "" || cfg.serverKey != "" {
		logger.Info(fmt.Sprintf("Rules engine service started using https on port %s with cert %s key %s",
			cfg.httpPort, cfg.serverCert, cfg.serverKey))
		errs <- http.ListenAndServeTLS(p, cfg.serverCert, cfg.serverKey, handler)
		return
	}
	logger.Info(fmt.Sprintf("Rules engine service started using http on port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, handler)
}

func createAuthClient(cfg config, tracer opentracing.Tracer, logger logger.Logger) (mainflux.AuthServiceClient, func() error) {
	if cfg.singleUserEmail != "" && cfg.singleUserToken != "" {
		return users.NewSingleUserService(cfg.singleUserEmail, cfg.singleUserToken), nil
	}

	conn := connectToGRPC(cfg.clientTLS, cfg.caCerts, cfg.authURL, logger)
	return authapi.NewClient(tracer, conn, cfg.authTimeout), conn.Close
}

func connectToGRPC(clientTLS bool, caCerts, URL string, logger logger.Logger) *grpc.ClientConn {
	var opts []grpc.DialOption
	if clientTLS {
		if caCerts != "" {
			tpc, err := credentials.NewClientTLSFromFile(caCerts, "")
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

	conn, err := grpc.Dial(URL, opts...)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to gRPC service: %s", err))
		os.Exit(1)
	}

	return conn
}
