package lora

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/nats-io/go-nats"

	apilora "github.com/brocaar/lora-app-server/api"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/logger"
)

const protocol = "lora"

// Service specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type Service interface {
	// Provision Lora App Server on MQTT broker
	ProvisionRouter(EventSourcing) error

	// Publish messages on Mainflux NATS broker
	MessageRouter(Message, *nats.Conn) error
}

var _ Service = (*adapterService)(nil)

type adapterService struct {
	natsConn      *nats.Conn
	loraAppClient apilora.ApplicationServiceClient
	logger        logger.Logger
	routeMap      RouteMapRepository
}

// EventSourcing is used to Unmarshal event sourcing data
type EventSourcing struct {
	CRUD string
	Type string
	ID   string
}

// New instantiates the HTTP adapter implementation.
func New(mc *nats.Conn, asc apilora.ApplicationServiceClient, m RouteMapRepository, logger logger.Logger) Service {
	return &adapterService{
		natsConn:      mc,
		loraAppClient: asc,
		routeMap:      m,
		logger:        logger,
	}
}

// ProvisionRouter routes provisioning from MainfluxNATS broker to Lora App Server gRPC API
func (as *adapterService) ProvisionRouter(provision EventSourcing) error {
	// TODO: do gRPC provisioning here if thing created
	switch provision.Type {
	case "app":
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		req := &apilora.CreateApplicationRequest{}
		as.loraAppClient.Create(ctx, req)

	default:
		as.logger.Error(fmt.Sprintf("Unknown provision type"))
		break
	}

	// TODO: save routeMap
	var channel uint64 = 1
	loraAppTopic := "application/123"
	if err := as.routeMap.Save(loraAppTopic, channel); err != nil {
		as.logger.Error(fmt.Sprintf("Failed to save route map: %s", err))
	}

	return nil
}

// MessageRouter routes messages from Lora MQTT broker to Mainflux NATS broker
func (as *adapterService) MessageRouter(m Message, nc *nats.Conn) error {
	eui, err := strconv.ParseUint(m.DevEUI, 10, 64)
	if err != nil {
		as.logger.Error(fmt.Sprintf("Failed to decode deviceEUI: %s", err.Error()))
		return nil
	}

	payload, err := base64.StdEncoding.DecodeString(m.Data)
	if err != nil {
		as.logger.Error(fmt.Sprintf("Failed to decode string message: %s", err.Error()))
		return nil
	}

	// Get route map of lora application
	channel, err := as.routeMap.Channel(m.ApplicationID)
	if err != nil {
		as.logger.Error(fmt.Sprintf("Routing doesn't exist for this LoRa application: %s", err.Error()))
		return nil
	}

	// Publish on Mainflux NATS broker
	msg := mainflux.RawMessage{
		Publisher:   eui,
		Protocol:    protocol,
		ContentType: "Content-Type",
		Channel:     channel,
		Payload:     payload,
	}

	data, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}

	subject := fmt.Sprintf("channel.%d", msg.Channel)
	return nc.Publish(subject, data)
}
