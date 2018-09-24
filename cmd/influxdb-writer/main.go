//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	kitprometheus "github.com/go-kit/kit/metrics/prometheus"
	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/writers"
	"github.com/mainflux/mainflux/writers/influxdb"
	"github.com/nats-io/go-nats"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
)

const (
	queue = "influxdb-writer"

	defNatsURL      = nats.DefaultURL
	defLogLevel     = "error"
	defPort         = "8180"
	defBatchSize    = "5000"
	defBatchTimeout = "5"
	defDBName       = "mainflux"
	defDBHost       = "localhost"
	defDBPort       = "8086"
	defDBUser       = "mainflux"
	defDBPass       = "mainflux"

	envNatsURL      = "MF_NATS_URL"
	envLogLevel     = "MF_INFLUX_WRITER_LOG_LEVEL"
	envPort         = "MF_INFLUX_WRITER_PORT"
	envBatchSize    = "MF_INFLUX_WRITER_BATCH_SIZE"
	envBatchTimeout = "MF_INFLUX_WRITER_BATCH_TIMEOUT"
	envDBName       = "MF_INFLUX_WRITER_DB_NAME"
	envDBHost       = "MF_INFLUX_WRITER_DB_HOST"
	envDBPort       = "MF_INFLUX_WRITER_DB_PORT"
	envDBUser       = "MF_INFLUX_WRITER_DB_USER"
	envDBPass       = "MF_INFLUX_WRITER_DB_PASS"
)

type config struct {
	NatsURL      string
	LogLevel     logger.Level
	Port         string
	BatchSize    string
	BatchTimeout string
	DBName       string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPass       string
}

func main() {
	cfg, clientCfg := loadConfigs()
	logger := logger.New(os.Stdout, cfg.LogLevel)

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to connect to NATS: %s", err))
		os.Exit(1)
	}
	defer nc.Close()

	client, err := influxdata.NewHTTPClient(clientCfg)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB client: %s", err))
		os.Exit(1)
	}
	defer client.Close()

	batchTimeout, batchSize, err := unmarshalInfluxDBSettings(cfg.BatchTimeout, cfg.BatchSize)
	if err != nil {
		logger.Error(err.Error())
	}

	timeout := time.Duration(batchTimeout) * time.Second
	repo, err := influxdb.New(client, cfg.DBName, batchSize, timeout)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create InfluxDB writer: %s", err))
		os.Exit(1)
	}

	counter, latency := makeMetrics()
	repo = writers.LoggingMiddleware(repo, logger)
	repo = writers.MetricsMiddleware(repo, counter, latency)
	if err := writers.Start(nc, repo, queue, logger); err != nil {
		logger.Error(fmt.Sprintf("Failed to start message writer: %s", err))
		os.Exit(1)
	}

	errs := make(chan error, 2)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go startHTTPService(cfg.Port, logger, errs)

	err = <-errs
	logger.Error(fmt.Sprintf("InfluxDB writer service terminated: %s", err))
}

func loadConfigs() (config, influxdata.HTTPConfig) {
	var logLevel logger.Level
	err := logLevel.UnmarshalText(mainflux.Env(envLogLevel, defLogLevel))
	if err != nil {
		log.Fatalf(`{"level":"error","message":"%s: %s","ts":"%s"}`, err, logLevel.String(), time.RFC3339Nano)
	}

	cfg := config{
		NatsURL:  mainflux.Env(envNatsURL, defNatsURL),
		LogLevel: logLevel,
		Port:     mainflux.Env(envPort, defPort),
		DBName:   mainflux.Env(envDBName, defDBName),
		DBHost:   mainflux.Env(envDBHost, defDBHost),
		DBPort:   mainflux.Env(envDBPort, defDBPort),
		DBUser:   mainflux.Env(envDBUser, defDBUser),
		DBPass:   mainflux.Env(envDBPass, defDBPass),
	}

	clientCfg := influxdata.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%s", cfg.DBHost, cfg.DBPort),
		Username: cfg.DBUser,
		Password: cfg.DBPass,
	}

	return cfg, clientCfg
}

func unmarshalInfluxDBSettings(cfgBatchTimeout string, cfgBatchSize string) (batchTimeout int, batchSize int, err error) {
	batchTimeout, err = strconv.Atoi(cfgBatchTimeout)
	if err != nil {
		return batchTimeout, batchSize, errors.New(fmt.Sprintf("Invalid value for batch timeout: %s", err))
	}

	batchSize, err = strconv.Atoi(cfgBatchSize)
	if err != nil {
		return batchTimeout, batchSize, errors.New(fmt.Sprintf("Invalid value of batch size: %s", err))
	}
	return batchTimeout, batchSize, err
}

func makeMetrics() (*kitprometheus.Counter, *kitprometheus.Summary) {
	counter := kitprometheus.NewCounterFrom(stdprometheus.CounterOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_count",
		Help:      "Number of database inserts.",
	}, []string{"method"})

	latency := kitprometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
		Namespace: "influxdb",
		Subsystem: "message_writer",
		Name:      "request_latency_microseconds",
		Help:      "Total duration of inserts in microseconds.",
	}, []string{"method"})

	return counter, latency
}

func startHTTPService(port string, logger logger.Logger, errs chan error) {
	p := fmt.Sprintf(":%s", port)
	logger.Info(fmt.Sprintf("InfluxDB writer service started, exposed port %s", p))
	errs <- http.ListenAndServe(p, influxdb.MakeHandler())
}
