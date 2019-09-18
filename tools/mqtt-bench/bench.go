// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package bench

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

const fixedSize = 411

// Benchmark - main benchmarking function
func Benchmark(cfg Config) {
	checkConnection(cfg.MQTT.Broker.URL, 1)

	var err error
	subsResults := map[string](*[]float64){}
	var caByte []byte
	if cfg.MQTT.TLS.MTLS {
		caFile, err := os.Open(cfg.MQTT.TLS.CA)
		defer caFile.Close()
		if err != nil {
			fmt.Println(err)
		}
		caByte, _ = ioutil.ReadAll(caFile)
	}

	mf := mainflux{}
	if _, err := toml.DecodeFile(cfg.Mf.ConnFile, &mf); err != nil {
		log.Fatalf("Cannot load Mainflux connections config %s \nUse tools/provision to create file", cfg.Mf.ConnFile)
	}

	resCh := make(chan *runResults)
	finishedPub := make(chan bool)

	startStamp := time.Now()

	n := len(mf.Channels)
	var cert tls.Certificate

	handler := getBytePayload(cfg.MQTT.Message.Size)

	start := time.Now()

	// Publishers
	for i := 0; i < cfg.Test.Pubs; i++ {
		mfChan := mf.Channels[i%n]
		mfThing := mf.Things[i%n]

		if cfg.MQTT.TLS.MTLS {
			cert, err = tls.X509KeyPair([]byte(mfThing.MTLSCert), []byte(mfThing.MTLSKey))
			if err != nil {
				log.Fatal(err)
			}
		}
		c := makeClient(i, cfg, mfChan, mfThing, startStamp, caByte, cert, handler)

		go c.publish(resCh)
	}

	// Collect the results
	var results []*runResults
	if cfg.Test.Pubs > 0 {
		results = make([]*runResults, cfg.Test.Pubs)
	}

	// Wait for publishers to finish
	go func() {
		for i := 0; i < cfg.Test.Pubs; i++ {
			results[i] = <-resCh
		}
		finishedPub <- true
	}()

	<-finishedPub

	totalTime := time.Now().Sub(start)
	totals := calculateTotalResults(results, totalTime, subsResults)
	if totals == nil {
		return
	}

	// Print sats
	printResults(results, totals, cfg.MQTT.Message.Format, cfg.Log.Quiet)
}

func getBytePayload(size int) handler {
	s := size - fixedSize
	if s <= 0 {
		s = 0
	}
	var b = make([]byte, s)
	rand.Read(b)

	return func(m *message) ([]byte, error) {
		m.Payload = b
		m.Sent = time.Now()
		return json.Marshal(m)
	}
}

func makeClient(i int, cfg Config, mfChan mfChannel, mfThing mfThing, start time.Time, caCert []byte, clientCert tls.Certificate, h handler) *Client {
	return &Client{
		ID:         strconv.Itoa(i),
		BrokerURL:  cfg.MQTT.Broker.URL,
		BrokerUser: mfThing.ThingID,
		BrokerPass: mfThing.ThingKey,
		MsgTopic:   fmt.Sprintf("channels/%s/messages/%d/test", mfChan.ChannelID, start.UnixNano()),
		MsgSize:    cfg.MQTT.Message.Size,
		MsgCount:   cfg.Test.Count,
		MsgQoS:     byte(cfg.MQTT.Message.QoS),
		Quiet:      cfg.Log.Quiet,
		MTLS:       cfg.MQTT.TLS.MTLS,
		SkipTLSVer: cfg.MQTT.TLS.SkipTLSVer,
		CA:         caCert,
		timeout:    cfg.MQTT.Timeout,
		ClientCert: clientCert,
		Retain:     cfg.MQTT.Message.Retain,
		SendMsg:    h,
	}
}
