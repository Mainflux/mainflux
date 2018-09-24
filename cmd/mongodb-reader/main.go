//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/readers"
	"github.com/mainflux/mainflux/readers/api"
	"github.com/mainflux/mainflux/readers/mongodb"
	thingsapi "github.com/mainflux/mainflux/things/api/grpc"
	"github.com/mongodb/mongo-go-driver/mongo"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

const (
	defThingsURL = "localhost:8181"
	defLogLevel  = "error"
	defPort      = "8180"
	defDBName    = "mainflux"
	defDBHost    = "localhost"
	defDBPort    = "27017"

	envThingsURL = "MF_THINGS_URL"
	envLogLevel  = "MF_MONGO_READER_LOG_LEVEL"
	envPort      = "MF_MONGO_READER_PORT"
	envDBName    = "MF_MONGO_READER_DB_NAME"
	envDBHost    = "MF_MONGO_READER_DB_HOST"
	envDBPort    = "MF_MONGO_READER_DB_PORT"
)

type config struct {
	thingsURL string
	logLevel  logger.Level
	port      string
	dbName    string
	dbHost    string
	dbPort    string
}

func main() {
	cfg := loadConfigs()
	logger := logger.New(os.Stdout, cfg.logLevel)

	conn := connectToThings(cfg.thingsURL, logger)
	defer conn.Close()

	tc := thingsapi.NewClient(conn)

	db := connectToMongoDB(cfg.dbHost, cfg.dbPort, cfg.dbName, logger)

	repo := newService(db, logger)

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPServer(repo, tc, cfg.port, logger, errs)

	err := <-errs
	logger.Error(fmt.Sprintf("MongoDB reader service terminated: %s", err))
}

func loadConfigs() config {
	var logLevel logger.Level
	err := logLevel.UnmarshalText(mainflux.Env(envLogLevel, defLogLevel))
	if err != nil {
		log.Fatalf(`{"level":"error","message":"%s: %s","ts":"%s"}`, err, logLevel.String(), time.RFC3339Nano)
	}

	return config{
		thingsURL: mainflux.Env(envThingsURL, defThingsURL),
		logLevel:  logLevel,
		port:      mainflux.Env(envPort, defPort),
		dbName:    mainflux.Env(envDBName, defDBName),
		dbHost:    mainflux.Env(envDBHost, defDBHost),
		dbPort:    mainflux.Env(envDBPort, defDBPort),
	}
}

func connectToMongoDB(host, port, name string, logger logger.Logger) *mongo.Database {
	client, err := mongo.Connect(context.Background(), fmt.Sprintf("mongodb://%s:%s", host, port), nil)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to database: %s", err))
		os.Exit(1)
	}

	return client.Database(name)
}

func connectToThings(url string, logger logger.Logger) *grpc.ClientConn {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to things service: %s", err))
		os.Exit(1)
	}

	return conn
}

func newService(db *mongo.Database, logger logger.Logger) readers.MessageRepository {
	repo := mongodb.New(db)
	repo = api.LoggingMiddleware(repo, logger)
	repo = api.MetricsMiddleware(
		repo,
		kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "mongodb",
			Subsystem: "message_reader",
			Name:      "request_count",
			Help:      "Number of requests received.",
		}, []string{"method"}),
		kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "mongodb",
			Subsystem: "message_reader",
			Name:      "request_latency_microseconds",
			Help:      "Total duration of requests in microseconds.",
		}, []string{"method"}),
	)

	return repo
}

func startHTTPServer(repo readers.MessageRepository, tc mainflux.ThingsServiceClient, port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("Mongo reader service started, exposed port %s", port))
	errs <- http.ListenAndServe(p, api.MakeHandler(repo, tc, "cassandra-reader"))
}
