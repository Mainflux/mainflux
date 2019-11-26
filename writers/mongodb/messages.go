// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/mainflux/mainflux/transformers/senml"
	"github.com/mainflux/mainflux/writers"
)

const collectionName string = "mainflux"

var _ writers.MessageRepository = (*mongoRepo)(nil)

type mongoRepo struct {
	db *mongo.Database
}

// Message struct is used as a MongoDB representation of Mainflux message.
type message struct {
	Channel     string   `bson:"channel,omitempty"`
	Subtopic    string   `bson:"subtopic,omitempty"`
	Publisher   string   `bson:"publisher,omitempty"`
	Protocol    string   `bson:"protocol,omitempty"`
	Name        string   `bson:"name,omitempty"`
	Unit        string   `bson:"unit,omitempty"`
	Value       *float64 `bson:"value,omitempty"`
	StringValue *string  `bson:"stringValue,omitempty"`
	BoolValue   *bool    `bson:"boolValue,omitempty"`
	DataValue   *string  `bson:"dataValue,omitempty"`
	Sum         *float64 `bson:"sum,omitempty"`
	Time        float64  `bson:"time,omitempty"`
	UpdateTime  float64  `bson:"updateTime,omitempty"`
	Link        string   `bson:"link,omitempty"`
}

// New returns new MongoDB writer.
func New(db *mongo.Database) writers.MessageRepository {
	return &mongoRepo{db}
}

func (repo *mongoRepo) Save(in ...interface{}) error {
	coll := repo.db.Collection(collectionName)
	msgs := []interface{}{}
	for _, msgIn := range in {
		msg, ok := msgIn.(senml.Message)
		if !ok {
			return errors.New("incorrect message type")
		}
		m := message{
			Channel:    msg.Channel,
			Subtopic:   msg.Subtopic,
			Publisher:  msg.Publisher,
			Protocol:   msg.Protocol,
			Name:       msg.Name,
			Unit:       msg.Unit,
			Time:       msg.Time,
			UpdateTime: msg.UpdateTime,
			Link:       msg.Link,
		}

		switch {
		case msg.Value != nil:
			m.Value = msg.Value
		case msg.StringValue != nil:
			m.StringValue = msg.StringValue
		case msg.DataValue != nil:
			m.DataValue = msg.DataValue
		case msg.BoolValue != nil:
			m.BoolValue = msg.BoolValue
		}
		m.Sum = msg.Sum

		msgs = append(msgs, m)
	}

	_, err := coll.InsertMany(context.Background(), msgs)
	return err
}
