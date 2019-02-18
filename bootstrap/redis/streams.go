package redis

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/mainflux/mainflux/bootstrap"
	"github.com/mainflux/mainflux/logger"
)

const (
	stream = "mainflux.things"
	group  = "mainflux.bootstrap"

	thingPrefix = "thing."
	thingRemove = thingPrefix + "remove"

	channelPrefix = "channel."
	channelUpdate = channelPrefix + "update"
	channelRemove = channelPrefix + "remove"
)

// EventStore represents event source for things and channels provisioning.
type EventStore interface {
	// Subscribes to given subject and receives events.
	Subscribe(string)
}

type eventStore struct {
	svc      bootstrap.Service
	client   *redis.Client
	consumer string
	logger   logger.Logger
}

// NewEventStore returns new event store instance.
func NewEventStore(svc bootstrap.Service, client *redis.Client, consumer string, log logger.Logger) EventStore {
	return eventStore{
		svc:      svc,
		client:   client,
		consumer: consumer,
		logger:   log,
	}
}

func (es eventStore) Subscribe(subject string) {
	es.client.XGroupCreateMkStream(stream, group, "$").Err()
	for {
		streams, err := es.client.XReadGroup(&redis.XReadGroupArgs{
			Group:    group,
			Consumer: es.consumer,
			Streams:  []string{stream, ">"},
			Count:    100,
		}).Result()
		if err != nil || len(streams) == 0 {
			continue
		}

		for _, msg := range streams[0].Messages {
			event := msg.Values

			var err error
			switch event["operation"] {
			case thingRemove:
				rte := decodeRemoveThing(event)
				err = es.handleRemoveThing(rte)
			case channelUpdate:
				uce := decodeUpdateChannel(event)
				err = es.handleUpdateChannel(uce)
			case channelRemove:
				rce := decodeRemoveChannel(event)
				err = es.handleRemoveChannel(rce)
			}
			if err != nil {
				es.logger.Warn(fmt.Sprintf("Failed to handle event sourcing: %s", err.Error()))
				break
			}
			es.client.XAck(stream, group, msg.ID)
		}
	}
}

func decodeRemoveThing(event map[string]interface{}) removeEvent {
	return removeEvent{
		id: read(event, "id", ""),
	}
}

func decodeUpdateChannel(event map[string]interface{}) updateChannelEvent {
	return updateChannelEvent{
		id:       read(event, "id", ""),
		name:     read(event, "name", ""),
		metadata: read(event, "metadata", ""),
	}
}

func decodeRemoveChannel(event map[string]interface{}) removeEvent {
	return removeEvent{
		id: read(event, "id", ""),
	}
}

func (es eventStore) handleRemoveThing(rte removeEvent) error {
	return es.svc.RemoveConfig(rte.id)
}

func (es eventStore) handleUpdateChannel(uce updateChannelEvent) error {
	channel := bootstrap.Channel{
		ID:       uce.id,
		Name:     uce.name,
		Metadata: uce.metadata,
	}
	return es.svc.UpdateChannel(channel)
}

func (es eventStore) handleRemoveChannel(rce removeEvent) error {
	println("remove chann")
	return nil
	// return es.svc.RemoveChannel(rce.id)
}

func read(event map[string]interface{}, key, def string) string {
	val, ok := event[key].(string)
	if !ok {
		return def
	}

	return val
}
