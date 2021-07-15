// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/things"
	"google.golang.org/grpc"
)

var _ mainflux.ThingsServiceClient = (*thingsClient)(nil)

// ServiceErrToken is used to simulate internal server error.
const ServiceErrToken = "unavailable"

type thingsClient struct {
	channels map[string]string // map[chanID]owner
}

// NewThingsClient returns mock implementation of things service client.
func NewThingsClient(channels map[string]string) mainflux.ThingsServiceClient {
	return &thingsClient{channels}
}

func (tc thingsClient) CanAccessByKey(ctx context.Context, req *mainflux.AccessByKeyReq, opts ...grpc.CallOption) (*mainflux.ThingID, error) {
	panic("not implemented")
}

func (tc thingsClient) CanAccessByID(context.Context, *mainflux.AccessByIDReq, ...grpc.CallOption) (*empty.Empty, error) {
	panic("not implemented")
}

func (tc thingsClient) IsChannelOwner(ctx context.Context, req *mainflux.ChannelOwnerReq, opts ...grpc.CallOption) (*empty.Empty, error) {
	owner, ok := tc.channels[req.ChanID]
	if !ok {
		return nil, things.ErrNotFound
	}

	if owner != req.GetOwner() {
		return nil, things.ErrNotFound
	}

	return nil, nil
}

func (tc thingsClient) Identify(ctx context.Context, req *mainflux.Token, opts ...grpc.CallOption) (*mainflux.ThingID, error) {
	panic("not implemented")
}
