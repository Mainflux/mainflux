//
// Copyright (c) 2019
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package twins_test

import (
	"context"
	"fmt"
	"testing"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mainflux/mainflux/twins"
	"github.com/mainflux/mainflux/twins/mocks"
	broker "github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	wrongID    = ""
	wrongValue = "wrong-value"
	email      = "user@example.com"
	token      = "token"
	natsURL    = "nats://localhost:4222"
	topic      = "topic"
)

var (
	twin = twins.Twin{Name: "test"}
)

func newService(tokens map[string]string) twins.Service {
	users := mocks.NewUsersService(tokens)
	twinsRepo := mocks.NewTwinRepository()
	idp := mocks.NewIdentityProvider()

	nc, _ := broker.Connect(natsURL)

	opts := mqtt.NewClientOptions()
	mc := mqtt.NewClient(opts)

	return twins.New(nc, mc, topic, users, twinsRepo, idp)
}

func TestAddTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})

	cases := []struct {
		desc  string
		twin  twins.Twin
		token string
		err   error
	}{
		{
			desc:  "add new twin",
			twin:  twins.Twin{Name: "a"},
			token: token,
			err:   nil,
		},
		{
			desc:  "add twin with wrong credentials",
			twin:  twins.Twin{Name: "d"},
			token: wrongValue,
			err:   twins.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range cases {
		_, err := svc.AddTwin(context.Background(), tc.token, tc.twin)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestUpdateTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	saved, err := svc.AddTwin(context.Background(), token, twin)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	other := twins.Twin{ID: wrongID, Key: "x"}

	cases := []struct {
		desc  string
		twin  twins.Twin
		token string
		err   error
	}{
		{
			desc:  "update existing twin",
			twin:  saved,
			token: token,
			err:   nil,
		},
		{
			desc:  "update twin with wrong credentials",
			twin:  saved,
			token: wrongValue,
			err:   twins.ErrUnauthorizedAccess,
		},
		{
			desc:  "update non-existing twin",
			twin:  other,
			token: token,
			err:   twins.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := svc.UpdateTwin(context.Background(), tc.token, tc.twin)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}

func TestViewTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	saved, err := svc.AddTwin(context.Background(), token, twin)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := map[string]struct {
		id    string
		token string
		err   error
	}{
		"view existing twin": {
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		"view twin with wrong credentials": {
			id:    saved.ID,
			token: wrongValue,
			err:   twins.ErrUnauthorizedAccess,
		},
		"view non-existing twin": {
			id:    wrongID,
			token: token,
			err:   twins.ErrNotFound,
		},
	}

	for desc, tc := range cases {
		_, err := svc.ViewTwin(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestListTwins(t *testing.T) {
	svc := newService(map[string]string{token: email})

	m := make(map[string]interface{})
	m["serial"] = "123456"
	twin.Metadata = m

	n := uint64(10)
	for i := uint64(0); i < n; i++ {
		svc.AddTwin(context.Background(), token, twin)
	}

	cases := map[string]struct {
		token    string
		offset   uint64
		limit    uint64
		name     string
		size     uint64
		metadata map[string]interface{}
		err      error
	}{
		"list all twins": {
			token: token,
			limit: n + 1,
			size:  n,
			err:   nil,
		},
		"list with zero limit": {
			token: token,
			limit: 0,
			size:  0,
			err:   nil,
		},
		"list with wrong credentials": {
			token: wrongValue,
			limit: 0,
			size:  0,
			err:   twins.ErrUnauthorizedAccess,
		},
		"list with metadata": {
			token:    token,
			limit:    n + 1,
			size:     n,
			err:      nil,
			metadata: m,
		},
	}

	for desc, tc := range cases {
		page, err := svc.ListTwins(context.Background(), tc.token, tc.limit, tc.name, tc.metadata)
		size := uint64(len(page.Twins))
		assert.Equal(t, tc.size, size, fmt.Sprintf("%s: expected %d got %d\n", desc, tc.size, size))
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", desc, tc.err, err))
	}
}

func TestRemoveTwin(t *testing.T) {
	svc := newService(map[string]string{token: email})
	saved, err := svc.AddTwin(context.Background(), token, twin)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s\n", err))

	cases := []struct {
		desc  string
		id    string
		token string
		err   error
	}{
		{
			desc:  "remove twin with wrong credentials",
			id:    saved.ID,
			token: wrongValue,
			err:   twins.ErrUnauthorizedAccess,
		},
		{
			desc:  "remove existing twin",
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove removed twin",
			id:    saved.ID,
			token: token,
			err:   nil,
		},
		{
			desc:  "remove non-existing twin",
			id:    wrongID,
			token: token,
			err:   nil,
		},
	}

	for _, tc := range cases {
		err := svc.RemoveTwin(context.Background(), tc.token, tc.id)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected %s got %s\n", tc.desc, tc.err, err))
	}
}
