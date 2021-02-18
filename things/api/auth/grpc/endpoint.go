// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/mainflux/mainflux/things"
)

func canAccessEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AccessByKeyReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		id, err := svc.CanAccessByKey(ctx, req.chanID, req.thingKey)
		if err != nil {
			return identityRes{}, err
		}
		return identityRes{id: id}, nil
	}
}

func canAccessByIDEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(accessByIDReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.CanAccessByID(ctx, req.chanID, req.thingID)
		return emptyRes{err: err}, err
	}
}

func canAccessChannelByOwnerEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(accessChannelByOwnerReq)
		if err := req.validate(); err != nil {
			return nil, err
		}

		err := svc.CanAccessChannelByOwner(ctx, req.chanID, req.owner)
		return emptyRes{err: err}, err
	}
}

func identifyEndpoint(svc things.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(identifyReq)
		id, err := svc.Identify(ctx, req.key)
		if err != nil {
			return identityRes{}, err
		}
		return identityRes{id: id}, nil
	}
}
