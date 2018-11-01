//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package cassandra_test

import (
	"fmt"
	"testing"

	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/writers/cassandra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const keyspace = "mainflux"

var (
	addr = "localhost"
	msg  = mainflux.Message{
		Channel:   1,
		Publisher: 1,
		Protocol:  "mqtt",
	}
	msgNum      = 42
	valueFields = 6
)

func TestSave(t *testing.T) {
	session, err := cassandra.Connect([]string{addr}, keyspace)
	require.Nil(t, err, fmt.Sprintf("failed to connect to Cassandra: %s", err))

	repo := cassandra.New(session)
	for i := 0; i < msgNum; i++ {
		count := i % valueFields
		switch count {
		case 0:
			msg.Values = &mainflux.Message_Value{5}
		case 1:
			msg.Values = &mainflux.Message_BoolValue{false}
		case 2:
			msg.Values = &mainflux.Message_StringValue{"value"}
		case 3:
			msg.Values = &mainflux.Message_DataValue{"base64data"}
		case 4:
			msg.ValueSum = nil
		case 5:
			msg.ValueSum = &mainflux.Sum{Value: 45}
		}

		err = repo.Save(msg)
		assert.Nil(t, err, fmt.Sprintf("expected no error, go %s", err))
	}
}
