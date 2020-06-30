// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/mainflux/mainflux"
	authapi "github.com/mainflux/mainflux/authn/api/grpc"
	"github.com/mainflux/mainflux/certs"
	"github.com/mainflux/mainflux/certs/api"
	"github.com/mainflux/mainflux/certs/postgres"
	"github.com/mainflux/mainflux/logger"
	"github.com/opentracing/opentracing-go"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/jmoiron/sqlx"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	mfsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	jconfig "github.com/uber/jaeger-client-go/config"
)

const (
	defLogLevel      = "error"
	defDBHost        = "localhost"
	defDBPort        = "5432"
	defDBUser        = "mainflux"
	defDBPass        = "mainflux"
	defDB            = "certs"
	defDBSSLMode     = "disable"
	defDBSSLCert     = ""
	defDBSSLKey      = ""
	defDBSSLRootCert = ""
	defClientTLS     = "false"
	defCACerts       = ""
	defPort          = "8204"
	defServerCert    = ""
	defServerKey     = ""
	defBaseURL       = "http://localhost"
	defThingsPrefix  = ""
	defJaegerURL     = ""
	defAuthnURL      = "localhost:8181"
	defAuthnTimeout  = "1s"

	defSignCAPath     = "ca.crt"
	defSignCAKeyPath  = "ca.key"
	defSignHoursValid = "2048h"
	defSignRSABits    = ""

	defVaultHost     = ""
	defVaultRole     = "mainflux"
	defVaultToken    = ""
	defVaultIssueURL = "pki_int/issue"

	envPort          = "MF_CERTS_HTTP_PORT"
	envLogLevel      = "MF_CERTS_LOG_LEVEL"
	envDBHost        = "MF_CERTS_DB_HOST"
	envDBPort        = "MF_CERTS_DB_PORT"
	envDBUser        = "MF_CERTS_DB_USER"
	envDBPass        = "MF_CERTS_DB_PASS"
	envDB            = "MF_CERTS_DB"
	envDBSSLMode     = "MF_CERTS_DB_SSL_MODE"
	envDBSSLCert     = "MF_CERTS_DB_SSL_CERT"
	envDBSSLKey      = "MF_CERTS_DB_SSL_KEY"
	envDBSSLRootCert = "MF_CERTS_DB_SSL_ROOT_CERT"
	envEncryptKey    = "MF_CERTS_ENCRYPT_KEY"
	envClientTLS     = "MF_CERTS_CLIENT_TLS"
	envCACerts       = "MF_CERTS_CA_CERTS"
	envServerCert    = "MF_CERTS_SERVER_CERT"
	envServerKey     = "MF_CERTS_SERVER_KEY"
	envBaseURL       = "MF_SDK_BASE_URL"
	envThingsPrefix  = "MF_SDK_THINGS_PREFIX"
	envJaegerURL     = "MF_JAEGER_URL"
	envAuthnURL      = "MF_AUTHN_GRPC_URL"
	envAuthnTimeout  = "MF_AUTHN_GRPC_TIMEOUT"

	envSignCAPath     = "MF_CERTS_SIGN_CA_PATH"
	envSignCAKey      = "MF_CERTS_SIGN_CA_KEY_PATH"
	envSignHoursValid = "MF_CERTS_SIGN_HOURS_VALID"
	envSignRSABits    = "MF_CERTS_SIGN_RSA_BITS"

	envVaultHost     = "MF_CERTS_VAULT_HOST"
	envVaultIssueURL = "MF_CERTS_VAULT_PKI"
	envVaultRole     = "MF_CERTS_VAULT_ROLE"
	envVaultToken    = "MF_CERTS_VAULT_TOKEN"
)

var (
	errFailedCertLoading         = errors.New("failed to load certificate")
	errFailedCertDecode          = errors.New("failed to decode certificate")
	errMissingCACertificate      = errors.New("missing CA")
	errPrivateKeyEmpty           = errors.New("private key empty")
	errPrivateKeyUnsupportedType = errors.New("private key unsupported type")
	errCertsRemove               = errors.New("failed to remove certificate")
	errCACertificateDoesntExist  = errors.New("CA certificate doesnt exist")
	errCAKeyDoesntExist          = errors.New("CA certificate key doesnt exist")
)

type config struct {
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
	jaegerURL      string
	authnURL       string
	authnTimeout   time.Duration
	signCAPath     string
	signCAKeyPath  string
	signRSABits    int
	signHoursValid string
	pkiIssueURL    string
	pkiAccessToken string
	pkiHost        string
	pkiRole        string
}

func main() {
	cfg := loadConfig()

	logger, err := mflog.New(os.Stdout, cfg.logLevel)
	if err != nil {
		log.Fatalf(err.Error())
	}

	tlsCert, caCert, err := loadCertificates(cfg)
	if err != nil {
		logger.Error("Failed to load CA certificates for issuing client certs")
	}

	pkiClient, err := initVaultClient(cfg)
	if err != nil {
		logger.Error("Failed to init vault client")
	}

	db := connectToDB(cfg.dbConfig, logger)
	defer db.Close()

	// esClient := connectToRedis(cfg.esURL, cfg.esPass, cfg.esDB, logger)
	// defer esClient.Close()

	authTracer, authCloser := initJaeger("auth", cfg.jaegerURL, logger)
	defer authCloser.Close()

	authConn := connectToAuth(cfg, logger)
	defer authConn.Close()

	auth := authapi.NewClient(authTracer, authConn, cfg.authnTimeout)

	svc := newService(auth, db, logger, nil, tlsCert, caCert, cfg, pkiClient)
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
	log.Println("Dbport:" + dbConfig.Port)
	log.Println("User:" + dbConfig.User)
	log.Println("Pass:" + dbConfig.Pass)

	authnTimeout, err := time.ParseDuration(mainflux.Env(envAuthnTimeout, defAuthnTimeout))
	if err != nil {
		log.Fatalf("Invalid %s value: %s", envAuthnTimeout, err.Error())
	}

	return config{
		logLevel:       mainflux.Env(envLogLevel, defLogLevel),
		dbConfig:       dbConfig,
		clientTLS:      tls,
		caCerts:        mainflux.Env(envCACerts, defCACerts),
		httpPort:       mainflux.Env(envPort, defPort),
		serverCert:     mainflux.Env(envServerCert, defServerCert),
		serverKey:      mainflux.Env(envServerKey, defServerKey),
		baseURL:        mainflux.Env(envBaseURL, defBaseURL),
		thingsPrefix:   mainflux.Env(envThingsPrefix, defThingsPrefix),
		jaegerURL:      mainflux.Env(envJaegerURL, defJaegerURL),
		authnURL:       mainflux.Env(envAuthnURL, defAuthnURL),
		authnTimeout:   authnTimeout,
		signCAKeyPath:  mainflux.Env(envSignCAKey, defSignCAKeyPath),
		signCAPath:     mainflux.Env(envSignCAPath, defSignCAPath),
		signHoursValid: mainflux.Env(envSignHoursValid, defSignHoursValid),
		pkiAccessToken: mainflux.Env(envVaultToken, defVaultToken),
		pkiIssueURL:    mainflux.Env(envVaultIssueURL, defVaultIssueURL),
		pkiRole:        mainflux.Env(envVaultRole, defVaultRole),
		pkiHost:        mainflux.Env(envVaultHost, defVaultHost),
	}

}
func initVaultClient(cfg config) (*vaultapi.Client, error) {
	conf := &vaultapi.Config{
		Address: cfg.pkiHost,
	}

	c, err := vaultapi.NewClient(conf)
	if err != nil {
		return nil, err
	}
	c.SetToken(cfg.pkiAccessToken)
	return c, nil

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

func newService(auth mainflux.AuthNServiceClient, db *sqlx.DB, logger mflog.Logger, esClient *redis.Client, tlsCert tls.Certificate, x509Cert *x509.Certificate, cfg config, pkiClient *vaultapi.Client) certs.Service {
	certsRepo := postgres.NewCertsRepository(db, logger)

	certsConfig := certs.Config{
		LogLevel:       cfg.logLevel,
		ClientTLS:      cfg.clientTLS,
		CaCerts:        cfg.caCerts,
		HttpPort:       cfg.httpPort,
		ServerCert:     cfg.serverCert,
		ServerKey:      cfg.serverKey,
		BaseURL:        cfg.baseURL,
		ThingsPrefix:   cfg.thingsPrefix,
		JaegerURL:      cfg.jaegerURL,
		AuthnURL:       cfg.authnURL,
		AuthnTimeout:   cfg.authnTimeout,
		SignTLSCert:    tlsCert,
		SignX509Cert:   x509Cert,
		SignHoursValid: cfg.signHoursValid,
		SignRSABits:    cfg.signRSABits,
		PkiAccessToken: cfg.pkiAccessToken,
		PkiHost:        cfg.pkiHost,
		PkiIssueURL:    cfg.pkiIssueURL,
		PkiRole:        cfg.pkiRole,
	}

	config := mfsdk.Config{
		BaseURL:      cfg.baseURL,
		ThingsPrefix: cfg.thingsPrefix,
	}

	sdk := mfsdk.NewSDK(config)

	svc := certs.New(auth, certsRepo, sdk, certsConfig, pkiClient)
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

func startHTTPServer(svc certs.Service, cfg config, logger mflog.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", cfg.httpPort)
	if cfg.serverCert != "" || cfg.serverKey != "" {
		logger.Info(fmt.Sprintf("Certs service started using https on port %s with cert %s key %s",
			cfg.httpPort, cfg.serverCert, cfg.serverKey))
		errs <- http.ListenAndServeTLS(p, cfg.serverCert, cfg.serverKey, api.MakeHandler(svc))
		return
	}
	logger.Info(fmt.Sprintf("Certs service started using http on port %s", cfg.httpPort))
	errs <- http.ListenAndServe(p, api.MakeHandler(svc))
}

func loadCertificates(conf config) (tls.Certificate, *x509.Certificate, error) {
	var tlsCert tls.Certificate
	var caCert *x509.Certificate

	if conf.signCAPath == "" || conf.signCAKeyPath == "" {
		return tlsCert, caCert, nil
	}

	if _, err := os.Stat(conf.signCAPath); os.IsNotExist(err) {
		return tlsCert, caCert, errCACertificateDoesntExist
	}

	if _, err := os.Stat(conf.signCAKeyPath); os.IsNotExist(err) {
		return tlsCert, caCert, errCAKeyDoesntExist
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

// func subscribeToThingsES(svc certs.Service, client *redis.Client, consumer string, logger mflog.Logger) {
// 	eventStore := rediscons.NewEventStore(svc, client, consumer, logger)
// 	logger.Info("Subscribed to Redis Event Store")
// 	if err := eventStore.Subscribe("mainflux.things"); err != nil {
// 		logger.Warn(fmt.Sprintf("Certs service failed to subscribe to event sourcing: %s", err))
// 	}
// }
