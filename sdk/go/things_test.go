//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package sdk_test

import (
	"fmt"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/mainflux/mainflux/sdk/go"
	"github.com/stretchr/testify/assert"

	"github.com/mainflux/mainflux/things"
	httpapi "github.com/mainflux/mainflux/things/api/http"
	"github.com/mainflux/mainflux/things/mocks"
)

const (
	contentType = "application/senml+json"
	email       = "user@example.com"
	otherEmail  = "other_user@example.com"
	token       = "token"
	otherToken  = "other_token"
	wrongValue  = "wrong_value"

	keyPrefix = "123e4567-e89b-12d3-a456-"
)

var (
	thing      = sdk.Thing{ID: "1", Type: "device", Name: "test_device", Metadata: "test_metadata"}
	emptyThing = sdk.Thing{}
)

func newThingsService(tokens map[string]string) things.Service {
	users := mocks.NewUsersService(tokens)
	thingsRepo := mocks.NewThingRepository()
	channelsRepo := mocks.NewChannelRepository(thingsRepo)
	chanCache := mocks.NewChannelCache()
	thingCache := mocks.NewThingCache()
	idp := mocks.NewIdentityProvider()

	return things.New(users, thingsRepo, channelsRepo, chanCache, thingCache, idp)
}

func newThingsServer(svc things.Service) *httptest.Server {
	mux := httpapi.MakeHandler(svc)
	return httptest.NewServer(mux)
}

func TestCreateThing(t *testing.T) {
	svc := newThingsService(map[string]string{token: email})
	ts := newThingsServer(svc)
	defer ts.Close()

	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)

	cases := []struct {
		desc     string
		thing    sdk.Thing
		token    string
		err      error
		location string
	}{
		{
			desc:     "create new thing",
			thing:    thing,
			token:    token,
			err:      nil,
			location: "/things/1",
		},
		{
			desc:     "create new thing with empty token",
			thing:    thing,
			token:    "",
			err:      sdk.ErrUnauthorized,
			location: "",
		},
		{
			desc:     "create new thing with invalid token",
			thing:    thing,
			token:    wrongValue,
			err:      sdk.ErrUnauthorized,
			location: "",
		},
		{
			desc:     "create new epmty thing",
			thing:    emptyThing,
			token:    wrongValue,
			err:      sdk.ErrInvalidArgs,
			location: "",
		},
	}
	for _, tc := range cases {
		loc, err := mainfluxSDK.CreateThing(tc.thing, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		assert.Equal(t, tc.location, loc, fmt.Sprintf("%s: expected location %s got %s", tc.desc, tc.location, loc))

	}
}

func TestThing(t *testing.T) {
	svc := newThingsService(map[string]string{token: email})
	ts := newThingsServer(svc)
	defer ts.Close()

	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	mainfluxSDK.CreateThing(thing, token)
	thing.Key = fmt.Sprintf("%s%012d", keyPrefix, 1)

	cases := []struct {
		desc     string
		thId     string
		token    string
		err      error
		response sdk.Thing
	}{
		{
			desc:     "Get existing thing",
			thId:     "1",
			token:    token,
			err:      nil,
			response: thing,
		},
		{
			desc:     "Get non-existent thing",
			thId:     "43",
			token:    token,
			err:      sdk.ErrNotFound,
			response: sdk.Thing{},
		},
		{
			desc:     "Get thing with invalid token",
			thId:     "1",
			token:    wrongValue,
			err:      sdk.ErrUnauthorized,
			response: sdk.Thing{},
		},
	}

	for _, tc := range cases {
		respTh, err := mainfluxSDK.Thing(tc.thId, tc.token)

		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, respTh, fmt.Sprintf("%s: expected response thing %s, got %s", tc.desc, tc.response, respTh))
	}

}

func TestThings(t *testing.T) {

	svc := newThingsService(map[string]string{token: email})
	ts := newThingsServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}
	var things []sdk.Thing

	mainfluxSDK := sdk.NewSDK(sdkConf)
	for i := 1; i < 101; i++ {

		th := sdk.Thing{ID: strconv.Itoa(i), Type: "device", Name: "test_device", Metadata: "test_metadata"}
		mainfluxSDK.CreateThing(th, token)
		th.Key = fmt.Sprintf("%s%012d", keyPrefix, i)
		things = append(things, th)
	}

	cases := []struct {
		desc     string
		token    string
		offset   uint64
		limit    uint64
		err      error
		response []sdk.Thing
	}{
		{
			desc:     "get a list of things",
			token:    token,
			offset:   0,
			limit:    5,
			err:      nil,
			response: things[0:5],
		},
		{
			desc:     "get a list of things with invalid token",
			token:    wrongValue,
			offset:   0,
			limit:    5,
			err:      sdk.ErrUnauthorized,
			response: nil,
		},
		{
			desc:     "get a list of things with empty token",
			token:    "",
			offset:   0,
			limit:    5,
			err:      sdk.ErrUnauthorized,
			response: nil,
		},
		{
			desc:     "get a list of things with zero limit",
			token:    token,
			offset:   0,
			limit:    0,
			err:      sdk.ErrInvalidArgs,
			response: nil,
		},
		{
			desc:     "get a list of things with limit greater than max",
			token:    token,
			offset:   0,
			limit:    110,
			err:      sdk.ErrInvalidArgs,
			response: nil,
		},
		{
			desc:     "get a list of things with offset greater than max",
			token:    token,
			offset:   110,
			limit:    5,
			err:      nil,
			response: nil,
		},
		{
			desc:     "get a list of things with invalid args (zero limit) and invalid token",
			token:    wrongValue,
			offset:   0,
			limit:    0,
			err:      sdk.ErrInvalidArgs,
			response: nil,
		},
	}
	for _, tc := range cases {
		respThs, err := mainfluxSDK.Things(tc.token, tc.offset, tc.limit)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
		assert.Equal(t, tc.response, respThs, fmt.Sprintf("%s: expected response channel %s, got %s", tc.desc, tc.response, respThs))

	}
}
func TestUpdateThing(t *testing.T) {
	svc := newThingsService(map[string]string{token: email})
	ts := newThingsServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	mainfluxSDK.CreateThing(thing, token)
	thing.Name = "test2"

	cases := []struct {
		desc  string
		thing sdk.Thing
		token string
		err   error
	}{
		{
			desc:  "update existing thing",
			thing: sdk.Thing{ID: "1", Type: "app", Name: "test_app", Metadata: "test_metadata2"},
			token: token,
			err:   nil,
		},
		{
			desc:  "update non-existing thing",
			thing: sdk.Thing{ID: "0", Type: "device", Name: "test_device", Metadata: "test_metadata"},
			token: token,
			err:   sdk.ErrNotFound,
		},
		{
			desc:  "update channel with invalid id",
			thing: sdk.Thing{ID: "invalid", Type: "device", Name: "test_device", Metadata: "test_metadata"},
			token: token,
			err:   sdk.ErrInvalidArgs,
		},
		{
			desc:  "update channel with invalid token",
			thing: sdk.Thing{ID: "1", Type: "app", Name: "test_app", Metadata: "test_metadata2"},
			token: wrongValue,
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "update channel with empty token",
			thing: sdk.Thing{ID: "1", Type: "app", Name: "test_app", Metadata: "test_metadata2"},
			token: "",
			err:   sdk.ErrUnauthorized,
		},
	}

	for _, tc := range cases {
		err := mainfluxSDK.UpdateThing(tc.thing, tc.token)

		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}

}

func TestDeleteThing(t *testing.T) {
	svc := newThingsService(map[string]string{token: email})
	ts := newThingsServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	mainfluxSDK.CreateThing(thing, token)

	cases := []struct {
		desc  string
		thId  string
		token string
		err   error
	}{

		{
			desc:  "delete thing with invalid token",
			thId:  "1",
			token: wrongValue,
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "delete non-existing thing",
			thId:  "2",
			token: token,
			err:   nil,
		},
		{
			desc:  "delete thing with invalid id",
			thId:  "invalid",
			token: token,
			err:   sdk.ErrFailedRemoval,
		},
		{
			desc:  "delete thing with empty token",
			thId:  "1",
			token: "",
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "delete existing thing",
			thId:  "1",
			token: token,
			err:   nil,
		},
		{
			desc:  "delete deleted thing",
			thId:  "1",
			token: token,
			err:   nil,
		},
	}

	for _, tc := range cases {
		err := mainfluxSDK.DeleteThing(tc.thId, tc.token)

		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestConnectThing(t *testing.T) {
	svc := newThingsService(map[string]string{
		token:      email,
		otherToken: otherEmail,
	})

	ts := newThingsServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	mainfluxSDK.CreateThing(thing, token)
	mainfluxSDK.CreateChannel(channel, token)
	mainfluxSDK.CreateChannel(channel, otherToken)

	cases := []struct {
		desc  string
		thId  string
		chId  string
		token string
		err   error
	}{

		{
			desc:  "connect existing thing to existing channel",
			thId:  "1",
			chId:  "1",
			token: token,
			err:   nil,
		},

		{
			desc:  "connect existing thing to non-existing channel",
			thId:  "1",
			chId:  "9",
			token: token,
			err:   sdk.ErrNotFound,
		},
		{
			desc:  "connect non-existing thing to existing channel",
			thId:  "9",
			chId:  "1",
			token: token,
			err:   sdk.ErrNotFound,
		},
		{
			desc:  "connect existing thing to channel with invalid ID",
			thId:  "1",
			chId:  "invalid",
			token: token,
			err:   sdk.ErrFailedConnection,
		},
		{
			desc:  "connect thing with invalid ID to existing channel",
			thId:  "invalid",
			chId:  "1",
			token: token,
			err:   sdk.ErrFailedConnection,
		},

		{
			desc:  "connect existing thing to existing channel with invalid token",
			thId:  "1",
			chId:  "1",
			token: wrongValue,
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "connect existing thing to existing channel with empty token",
			thId:  "1",
			chId:  "1",
			token: "",
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "connect thing from owner to channel of other user",
			thId:  "1",
			chId:  "2",
			token: token,
			err:   sdk.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := mainfluxSDK.ConnectThing(tc.thId, tc.chId, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}

func TestDisconnectThing(t *testing.T) {
	svc := newThingsService(map[string]string{
		token:      email,
		otherToken: otherEmail,
	})

	ts := newThingsServer(svc)
	defer ts.Close()
	sdkConf := sdk.Config{
		BaseURL:           ts.URL,
		UsersPrefix:       "",
		ThingsPrefix:      "",
		HTTPAdapterPrefix: "",
		MsgContentType:    contentType,
		TLSVerification:   false,
	}

	mainfluxSDK := sdk.NewSDK(sdkConf)
	mainfluxSDK.CreateThing(thing, token)
	mainfluxSDK.CreateChannel(channel, token)
	mainfluxSDK.ConnectThing("1", "1", token)
	mainfluxSDK.CreateChannel(channel, otherToken)

	cases := []struct {
		desc  string
		thId  string
		chId  string
		token string
		err   error
	}{

		{
			desc:  "disconnect connected thing from channel",
			thId:  "1",
			chId:  "1",
			token: token,
			err:   nil,
		},

		{
			desc:  "disconnect existing thing from non-existing channel",
			thId:  "1",
			chId:  "9",
			token: token,
			err:   sdk.ErrNotFound,
		},
		{
			desc:  "disconnect non-existing thing from existing channel",
			thId:  "9",
			chId:  "1",
			token: token,
			err:   sdk.ErrNotFound,
		},
		{
			desc:  "disconnect existing thing from channel with invalid ID",
			thId:  "1",
			chId:  "invalid",
			token: token,
			err:   sdk.ErrFailedConnection,
		},
		{
			desc:  "disconnect thing with invalid ID from existing channel",
			thId:  "invalid",
			chId:  "1",
			token: token,
			err:   sdk.ErrFailedConnection,
		},

		{
			desc:  "disconnect existing thing from existing channel with invalid token",
			thId:  "1",
			chId:  "1",
			token: wrongValue,
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "disconnect existing thing from existing channel with empty token",
			thId:  "1",
			chId:  "1",
			token: "",
			err:   sdk.ErrUnauthorized,
		},
		{
			desc:  "disconnect owner's thing from someone elses channel",
			thId:  "1",
			chId:  "2",
			token: token,
			err:   sdk.ErrNotFound,
		},
	}

	for _, tc := range cases {
		err := mainfluxSDK.DisconnectThing(tc.thId, tc.chId, tc.token)
		assert.Equal(t, tc.err, err, fmt.Sprintf("%s: expected error %s, got %s", tc.desc, tc.err, err))
	}
}
