package main

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/GaryBoone/GoStats/stats"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// SubTimes - measuring time of arrival of message in subs
type SubTimes map[string][]float64

// Client - represents mqtt client
type Client struct {
	ID         string
	BrokerURL  string
	BrokerUser string
	BrokerPass string
	MsgTopic   string
	MsgSize    int
	MsgCount   int
	MsgQoS     byte
	Quiet      bool
	mqttClient *mqtt.Client
	Mtls       bool
	SkipTlsVer bool
	CA         []byte
	clientCert tls.Certificate
	ClientKey  *rsa.PrivateKey
}

// RunPublisher - runs publisher
func (c *Client) RunPublisher(res chan *RunResults, mtls bool) {
	newMsgs := make(chan *Message)
	pubMsgs := make(chan *Message)
	doneGen := make(chan bool)
	donePub := make(chan bool)
	runResults := new(RunResults)

	started := time.Now()
	// Start generator
	go c.genMessages(newMsgs, doneGen)
	// Start publisher
	go c.pubMessages(newMsgs, pubMsgs, doneGen, donePub, mtls)

	times := []float64{}
	for {
		select {
		case m := <-pubMsgs:
			cid := m.ID
			if m.Error {
				runResults.Failures++
			} else {
				runResults.Successes++
				runResults.ID = cid
				times = append(times, float64(m.Delivered.Sub(m.Sent).Nanoseconds()/1000)) // in microseconds
			}
		case <-donePub:
			// Calculate results
			duration := time.Now().Sub(started)
			runResults.MsgTimeMin = stats.StatsMin(times)
			runResults.MsgTimeMax = stats.StatsMax(times)
			runResults.MsgTimeMean = stats.StatsMean(times)
			runResults.MsgTimeStd = stats.StatsSampleStandardDeviation(times)
			runResults.RunTime = duration.Seconds()
			runResults.MsgsPerSec = float64(runResults.Successes) / duration.Seconds()

			// Report results and exit
			res <- runResults
			return
		}
	}
}

// RunSubscriber - runs a subscriber
func (c *Client) RunSubscriber(wg *sync.WaitGroup, subTimes *SubTimes, done *chan bool, mtls bool) {
	defer wg.Done()
	// Start subscriber
	c.subscribe(wg, subTimes, done, mtls)

}

func (c *Client) genMessages(ch chan *Message, done chan bool) {
	for i := 0; i < c.MsgCount; i++ {
		msgPayload := MessagePayload{Payload: make([]byte, c.MsgSize)}
		ch <- &Message{
			Topic:   c.MsgTopic,
			QoS:     c.MsgQoS,
			Payload: msgPayload,
		}
	}
	done <- true
	return
}

func (c *Client) subscribe(wg *sync.WaitGroup, subTimes *SubTimes, done *chan bool, mtls bool) {
	clientID := fmt.Sprintf("sub-%v-%v", time.Now().Format(time.RFC3339Nano), c.ID)
	c.ID = clientID

	onConnected := func(client mqtt.Client) {
		if !c.Quiet {
			log.Printf("CLIENT %v is connected to the broker %v\n", clientID, c.BrokerURL)
		}
	}

	c.connect(onConnected, mtls)

	token := (*c.mqttClient).Subscribe(c.MsgTopic, 0, func(cl mqtt.Client, msg mqtt.Message) {

		mp := MessagePayload{}
		err := json.Unmarshal(msg.Payload(), &mp)
		if err != nil {
			log.Printf("CLIENT %s failed to decode message", clientID)
		}
	})

	token.Wait()

}

func (c *Client) pubMessages(in, out chan *Message, doneGen chan bool, donePub chan bool, mtls bool) {

	clientID := fmt.Sprintf("pub-%v-%v", time.Now().Format(time.RFC3339Nano), c.ID)
	c.ID = clientID
	onConnected := func(client mqtt.Client) {
		if !c.Quiet {
			log.Printf("CLIENT %v is connected to the broker %v\n", clientID, c.BrokerURL)
		}
		ctr := 0
		for {
			select {
			case m := <-in:
				m.Sent = time.Now()
				m.ID = clientID
				m.Payload.ID = clientID
				m.Payload.Sent = m.Sent

				pload, _ := json.Marshal(m.Payload)
				token := client.Publish(m.Topic, m.QoS, false, pload)
				token.Wait()
				if token.Error() != nil {
					m.Error = true
				} else {
					m.Delivered = time.Now()
					m.Error = false
				}
				out <- m

				if ctr > 0 && ctr%100 == 0 {
					if !c.Quiet {
						log.Printf("CLIENT %v published %v messages and keeps publishing...\n", clientID, ctr)
					}
				}
				ctr++
			case <-doneGen:
				donePub <- true
				if !c.Quiet {
					log.Printf("CLIENT %v is done publishing\n", clientID)
				}
				return
			}
		}
	}

	c.connect(onConnected, mtls)

}

func (c *Client) connect(onConnected func(client mqtt.Client), mtls bool) error {
	opts := mqtt.NewClientOptions().
		AddBroker(c.BrokerURL).
		SetClientID(c.ID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetOnConnectHandler(onConnected).
		SetConnectionLostHandler(func(client mqtt.Client, reason error) {
			log.Printf("CLIENT %s lost connection to the broker: %v. Will reconnect...\n", c.ID, reason.Error())
		})
	if c.BrokerUser != "" && c.BrokerPass != "" {
		opts.SetUsername(c.BrokerUser)
		opts.SetPassword(c.BrokerPass)
	}
	if mtls {

		cfg := &tls.Config{
			InsecureSkipVerify: c.SkipTlsVer,
		}

		if c.CA != nil {
			cfg.RootCAs = x509.NewCertPool()
			if cfg.RootCAs.AppendCertsFromPEM(c.CA) {
				log.Printf("Successfully added certificate\n")
			}
		}
		if c.clientCert.Certificate != nil {
			cfg.Certificates = []tls.Certificate{c.clientCert}
		}

		cfg.BuildNameToCertificate()
		opts.SetTLSConfig(cfg)
		opts.SetProtocolVersion(4)
	}
	client := mqtt.NewClient(opts)
	token := client.Connect()
	token.Wait()

	c.mqttClient = &client

	if token.Error() != nil {
		log.Printf("CLIENT %v had error connecting to the broker: %s\n", c.ID, token.Error())
		return token.Error()
	}
	return nil
}
