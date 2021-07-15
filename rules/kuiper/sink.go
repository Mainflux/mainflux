// +build kuiper

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/emqx/kuiper/common"
	"github.com/emqx/kuiper/xstream/api"
	"github.com/mainflux/mainflux/pkg/messaging"
	"github.com/mainflux/mainflux/pkg/messaging/nats"
	"github.com/mainflux/senml"
)

const baseName = "rules_engine"

type mainfluxConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Channel  string `json:"channel"`
	Subtopic string `json:"subtopic"`
}

type mainfluxSink struct {
	cfg *mainfluxConfig
	pub nats.Publisher
}

// Configure sink with properties from rule action definition
func (ms *mainfluxSink) Configure(props map[string]interface{}) error {
	cfg := &mainfluxConfig{}

	if err := common.MapToStruct(props, cfg); err != nil {
		return fmt.Errorf("Read properties %v fail with error: %v", props, err)
	}
	if cfg.Host == "" {
		return fmt.Errorf("property Host is required")
	}
	if cfg.Port == "" {
		return fmt.Errorf("property Port is required")
	}
	if cfg.Channel == "" {
		return fmt.Errorf("property Channel is required")
	}

	ms.cfg = cfg

	return nil
}

func (ms *mainfluxSink) Open(ctx api.StreamContext) (err error) {
	logger := ctx.GetLogger()
	logger.Debug("Opening mainflux sink")

	addr := fmt.Sprintf("%s:%s/", ms.cfg.Host, ms.cfg.Port)
	pub, err := nats.NewPublisher(addr)
	if err != nil {
		return fmt.Errorf("Failed to connect to nats at address %s with error: %v", addr, err)
	}
	ms.pub = pub

	return
}

// Collect publishes messages to Mainflux NATS broker
func (ms *mainfluxSink) Collect(ctx api.StreamContext, item interface{}) error {
	logger := ctx.GetLogger()
	logger.Debugf("mainflux sink receive %v", item)

	var msg messaging.Message
	msg.Channel = ms.cfg.Channel
	msg.Subtopic = ms.cfg.Subtopic
	msg.Created = time.Now().Unix()

	itemBytes, ok := item.([]byte)
	if !ok {
		logger.Debug("mainflux sink received non byte data")
	}
	var rec []senml.Record
	if err := json.Unmarshal(itemBytes, &rec); err != nil {
		return fmt.Errorf("Failed to unmarshal %v to senml", item)
	}
	if len(rec) == 0 {
		return errors.New("Empty senml array")
	}
	if rec[0].BaseName == "" {
		rec[0].BaseName = baseName
	}
	pack := senml.Pack{Records: []senml.Record{rec[0]}}
	payload, err := senml.Encode(pack, senml.JSON)
	if err != nil {
		return fmt.Errorf("Failed to encode %v to JSON", pack)
	}

	msg.Payload = payload
	if err := ms.pub.Publish(ms.cfg.Channel, msg); err != nil {
		return fmt.Errorf("Failed to publish message to %s", ms.cfg.Channel)
	}

	return nil
}

func (ms *mainfluxSink) Close(ctx api.StreamContext) error {
	if ms.pub != nil {
		ms.pub.Close()
	}

	return nil
}

// Mainflux exports the constructor used to instantiates the plugin
func Mainflux() api.Sink {
	return &mainfluxSink{}
}
