// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package grpc_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	"github.com/mainflux/mainflux/authn"
	"github.com/mainflux/mainflux/authz"
	grpcapi "github.com/mainflux/mainflux/authz/api/grpc"
	pb "github.com/mainflux/mainflux/authz/api/pb"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	port   = 8081
	secret = "secret"
	email  = "test@example.com"
)

var svc authz.Service

func newService() authz.Service {
	m := model.NewModel()
	m.AddDef("r", "r", "sub, obj, act")
	m.AddDef("p", "p", "sub, obj, act")
	m.AddDef("g", "g", "_, _")
	m.AddDef("e", "e", "some(where (p.eft == allow))")
	m.AddDef("m", "m", "( g(r.sub, p.sub) && keyMatch(r.obj, p.obj) && regexMatch(r.act, p.act) ) || r.sub == 'admin@example.com'")
	e, _ := casbin.NewSyncedEnforcer(m)

	_, _ = e.AddPolicy("admin", "data1", "read")

	return authz.New(e)
}

func startGRPCServer(svc authz.Service, port int) {
	listener, _ := net.Listen("tcp", fmt.Sprintf(":%d", port))
	server := grpc.NewServer()
	pb.RegisterAuthZServiceServer(server, grpcapi.NewServer(mocktracer.New(), svc))
	go server.Serve(listener)
}

func TestAuthorize(t *testing.T) {

	authAddr := fmt.Sprintf("localhost:%d", port)
	conn, err := grpc.Dial(authAddr, grpc.WithInsecure())
	require.Nil(t, err, fmt.Sprintf("user id unexpected error: %s", err))
	client := grpcapi.NewClient(conn, mocktracer.New(), time.Second)

	cases := []struct {
		desc    string
		id      string
		kind    uint32
		err     error
		code    codes.Code
		request pb.AuthorizeReq
	}{
		{
			desc:    "authorize with correct request",
			id:      email,
			kind:    authn.UserKey,
			err:     nil,
			code:    codes.OK,
			request: pb.AuthorizeReq{Sub: "admin", Obj: "data1", Act: "read"},
		},
		{
			desc:    "authorize with incorrect request",
			id:      email,
			kind:    authn.UserKey,
			err:     nil,
			code:    codes.OK,
			request: pb.AuthorizeReq{Sub: "admin", Obj: "data1", Act: "read"},
		},
	}

	for _, tc := range cases {
		_, err := client.Authorize(context.Background(), &tc.request)
		e, ok := status.FromError(err)
		assert.True(t, ok, "gRPC status can't be extracted from the error")
		assert.Equal(t, tc.code, e.Code(), fmt.Sprintf("%s: expected %s got %s", tc.desc, tc.code, e.Code()))
	}
}
