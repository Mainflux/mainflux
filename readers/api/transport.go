// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/gorilla/schema"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/readers"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	contentType    = "application/json"
	offsetKey      = "offset"
	limitKey       = "limit"
	formatKey      = "format"
	subtopicKey    = "subtopic"
	publisherKey   = "publisher"
	protocolKey    = "protocol"
	nameKey        = "name"
	valueKey       = "v"
	stringValueKey = "vs"
	dataValueKey   = "vd"
	comparatorKey  = "comparator"
	fromKey        = "from"
	toKey          = "to"
	defLimit       = 10
	defOffset      = 0
	defFormat      = "messages"
)

var listKeys = []string{
	"limit",
	"format",
	"subtopic",
	"publisher",
	"protocol",
	"name",
	"v",
	"vs",
	"vd",
	"comparator",
	"from",
	"to",
}

var (
	errUnauthorizedAccess = errors.New("missing or invalid credentials provided")
	auth                  mainflux.ThingsServiceClient
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(svc readers.MessageRepository, tc mainflux.ThingsServiceClient, svcName string) http.Handler {
	auth = tc

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	mux := bone.New()
	mux.Get("/channels/:chanID/messages", kithttp.NewServer(
		listMessagesEndpoint(svc),
		decodeList,
		encodeResponse,
		opts...,
	))
	mux.Post("/channels/:chanID/messages/search", kithttp.NewServer(
		listMessagesEndpoint(svc),
		decodeSearch,
		encodeResponse,
		opts...,
	))
	mux.GetFunc("/version", mainflux.Version(svcName))
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func decodeList(_ context.Context, r *http.Request) (interface{}, error) {
	fmt.Println("Decoding list...")
	chanID := bone.GetValue(r, "chanID")
	if chanID == "" {
		return nil, errors.ErrInvalidQueryParams
	}

	if err := authorize(r, chanID); err != nil {
		return nil, err
	}

	var q query
	if err := schema.NewDecoder().Decode(&q, r.URL.Query()); err != nil {
		fmt.Println("ERROR DECODE PARMAS:", err)
		return nil, err
	}
	q.ChannelID = chanID
	if q.Format == "" {
		q.Format = defFormat
	}
	if q.Limit == 0 {
		q.Limit = defLimit
	}

	meta, err := q.toPageMeta()
	if err != nil {
		fmt.Println("ERROR DECODE meta", err)
		return nil, err
	}
	// meta.ChanID = chanID
	req := listMessagesReq{
		pageMeta: meta,
	}

	return req, nil
}

func decodeSearch(_ context.Context, r *http.Request) (interface{}, error) {
	chanID := bone.GetValue(r, "chanID")
	if chanID == "" {
		return nil, errors.ErrInvalidQueryParams
	}
	if err := authorize(r, chanID); err != nil {
		return nil, err
	}

	var pm readers.PageMetadata
	if err := json.NewDecoder(r.Body).Decode(&pm); err != nil {
		return nil, err
	}
	// pm.ChanID = chanID
	pm["channel"] = chanID
	req := listMessagesReq{
		pm,
	}
	fmt.Println("search:", req)

	return req, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", contentType)

	if ar, ok := response.(mainflux.Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}
		w.WriteHeader(ar.Code())
		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch {
	case errors.Contains(err, nil):
	case errors.Contains(err, errors.ErrInvalidQueryParams):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Contains(err, errUnauthorizedAccess):
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	errorVal, ok := err.(errors.Error)
	if ok {
		w.Header().Set("Content-Type", contentType)
		if err := json.NewEncoder(w).Encode(errorRes{Err: errorVal.Msg()}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

func authorize(r *http.Request, chanID string) error {
	token := r.Header.Get("Authorization")
	if token == "" {
		return errUnauthorizedAccess
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := auth.CanAccessByKey(ctx, &mainflux.AccessByKeyReq{Token: token, ChanID: chanID})
	if err != nil {
		e, ok := status.FromError(err)
		if ok && e.Code() == codes.PermissionDenied {
			return errUnauthorizedAccess
		}
		return err
	}

	return nil
}

func readBoolValueQuery(r *http.Request, key string) (bool, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return false, errors.ErrInvalidQueryParams
	}

	if len(vals) == 0 {
		return false, errors.ErrNotFoundParam
	}

	b, err := strconv.ParseBool(vals[0])
	if err != nil {
		return false, errors.ErrInvalidQueryParams
	}

	return b, nil
}

func readQueryParam(r *http.Request, key string) (string, error) {
	vals := bone.GetQuery(r, key)
	if len(vals) > 1 {
		return "", errors.ErrInvalidQueryParams
	}

	if len(vals) == 0 {
		return "", nil
	}

	return vals[0], nil
}
