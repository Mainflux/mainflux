// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/mainflux/mainflux/pkg/errors"

	kitot "github.com/go-kit/kit/tracing/opentracing"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/rules"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	contentType = "application/json"
	name        = "name"
	action      = "action"
	kuiperType  = "kuiperType"
)

var (
	errUnsupportedContentType = errors.New("unsupported content type")
	errInvalidQueryParams     = errors.New("invalid query params")
)

// MakeHandler returns a HTTP handler for API endpoints.
func MakeHandler(tracer opentracing.Tracer, svc rules.Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	r := bone.New()

	r.Get("/info", kithttp.NewServer(
		kitot.TraceServer(tracer, "info")(infoEndpoint(svc)),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Post("/streams", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_stream")(createStreamEndpoint(svc)),
		decodeCreateStream,
		encodeResponse,
		opts...,
	))

	r.Put("/streams/:name", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_stream")(updateStreamEndpoint(svc)),
		decodeUpdateStream,
		encodeResponse,
		opts...,
	))

	r.Get("/streams", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_streams")(listStreamsEndpoint(svc)),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Get("/streams/:name", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_stream")(viewStreamEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Delete("/:"+kuiperType+"/:name", kithttp.NewServer(
		kitot.TraceServer(tracer, "delete")(deleteEndpoint(svc)),
		decodeDelete,
		encodeResponse,
		opts...,
	))

	r.Post("/rules", kithttp.NewServer(
		kitot.TraceServer(tracer, "create_rule")(createRuleEndpoint(svc)),
		decodeCreateRule,
		encodeResponse,
		opts...,
	))

	r.Put("/rules/:name", kithttp.NewServer(
		kitot.TraceServer(tracer, "update_rule")(updateRuleEndpoint(svc)),
		decodeUpdateRule,
		encodeResponse,
		opts...,
	))

	r.Get("/rules", kithttp.NewServer(
		kitot.TraceServer(tracer, "list_rules")(listRulesEndpoint(svc)),
		decodeList,
		encodeResponse,
		opts...,
	))

	r.Get("/rules/:name", kithttp.NewServer(
		kitot.TraceServer(tracer, "view_rule")(viewRuleEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))
	r.Get("/rules/:name/status", kithttp.NewServer(
		kitot.TraceServer(tracer, "rule_status")(ruleStatusEndpoint(svc)),
		decodeView,
		encodeResponse,
		opts...,
	))

	r.Post("/rules/:name/:action", kithttp.NewServer(
		kitot.TraceServer(tracer, "control_rule")(controlRuleEndpoint(svc)),
		decodeControl,
		encodeResponse,
		opts...,
	))

	r.GetFunc("/version", mainflux.Version("things"))
	r.Handle("/metrics", promhttp.Handler())

	return r
}

func decodeList(_ context.Context, r *http.Request) (interface{}, error) {
	req := listReq{
		token: r.Header.Get("Authorization"),
	}
	return req, nil
}

func decodeCreateStream(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := streamReq{
		token: r.Header.Get("Authorization"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req.stream); err != nil {
		return nil, errors.Wrap(rules.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateStream(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := streamReq{
		token: r.Header.Get("Authorization"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req.stream); err != nil {
		return nil, errors.Wrap(rules.ErrMalformedEntity, err)
	}

	req.stream.Name = bone.GetValue(r, name)

	return req, nil
}

func decodeView(_ context.Context, r *http.Request) (interface{}, error) {
	req := viewReq{
		token: r.Header.Get("Authorization"),
		name:  bone.GetValue(r, name),
	}
	return req, nil
}

func decodeCreateRule(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := ruleReq{
		token: r.Header.Get("Authorization"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(rules.ErrMalformedEntity, err)
	}

	return req, nil
}

func decodeUpdateRule(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), contentType) {
		return nil, errUnsupportedContentType
	}

	req := ruleReq{
		token: r.Header.Get("Authorization"),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, errors.Wrap(rules.ErrMalformedEntity, err)
	}

	req.ID = bone.GetValue(r, name)

	return req, nil
}

func decodeControl(_ context.Context, r *http.Request) (interface{}, error) {
	req := controlReq{
		token:  r.Header.Get("Authorization"),
		id:     bone.GetValue(r, name),
		action: bone.GetValue(r, action),
	}
	return req, nil
}

func decodeDelete(_ context.Context, r *http.Request) (interface{}, error) {
	req := deleteReq{
		token:      r.Header.Get("Authorization"),
		name:       bone.GetValue(r, name),
		kuiperType: bone.GetValue(r, kuiperType),
	}
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
	w.Header().Set("Content-Type", contentType)

	switch errorVal := err.(type) {
	case errors.Error:
		w.Header().Set("Content-Type", contentType)
		switch {
		case errors.Contains(errorVal, rules.ErrUnauthorizedAccess):
			w.WriteHeader(http.StatusUnauthorized)

		case errors.Contains(errorVal, rules.ErrNotFound):
			w.WriteHeader(http.StatusNotFound)

		case errors.Contains(errorVal, rules.ErrMalformedEntity),
			errors.Contains(errorVal, rules.ErrKuiperServer),
			errors.Contains(errorVal, errInvalidQueryParams):
			w.WriteHeader(http.StatusBadRequest)

		case errors.Contains(errorVal, errUnsupportedContentType):
			w.WriteHeader(http.StatusUnsupportedMediaType)

		case errors.Contains(errorVal, io.ErrUnexpectedEOF),
			errors.Contains(errorVal, io.EOF):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		if errorVal.Msg() != "" {
			errTxt := errorVal.Msg()
			if errorVal.Err() != nil {
				errTxt += " : " + errorVal.Err().Msg()
			}
			if err := json.NewEncoder(w).Encode(errorRes{Err: errTxt}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
