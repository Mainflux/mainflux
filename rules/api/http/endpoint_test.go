// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mainflux/mainflux/rules"
	httpapi "github.com/mainflux/mainflux/rules/api/http"
	"github.com/mainflux/mainflux/rules/mocks"
	"github.com/opentracing/opentracing-go/mocktracer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	url         = "localhost"
	contentType = "application/json"
	token       = "token"
	token2      = "token2"
	wrong       = "wrong"
	email       = "angry_albattani@email.com"
	email2      = "xenodochial_goldwasser@email.com"
	channel     = "103ec2f2-2034-4d9e-8039-13f4efd36b04"
	channel2    = "243fec72-7cf7-4bca-ac87-44a53b318510"
	action      = "start"
	name        = "name"
	name2       = "name2"
	row         = "v float, n string"
	sql         = "select * from stream where v > 1.2;"
)

var (
	stream = rules.Stream{
		Name:    name,
		Channel: channel,
		Row:     row,
		Host:    url,
	}
	stream2 = rules.Stream{
		Name:    name2,
		Channel: channel2,
		Row:     row,
		Host:    url,
	}
	rule  = mocks.CreateRule("rule", channel)
	rule2 = mocks.CreateRule("rule2", channel2)
)

type testRequest struct {
	client      *http.Client
	method      string
	url         string
	contentType string
	token       string
	body        io.Reader
}

func (tr testRequest) make() (*http.Response, error) {
	req, err := http.NewRequest(tr.method, tr.url, tr.body)
	if err != nil {
		return nil, err
	}
	if tr.token != "" {
		req.Header.Set("Authorization", tr.token)
	}
	if tr.contentType != "" {
		req.Header.Set("Content-Type", tr.contentType)
	}
	return tr.client.Do(req)
}

func newServer(svc rules.Service) *httptest.Server {
	mux := httpapi.MakeHandler(mocktracer.New(), svc)
	return httptest.NewServer(mux)
}

func toJSON(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

func TestCreateStream(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	ts := newServer(svc)
	defer ts.Close()

	valid := toJSON(stream)

	invalidStream := stream
	invalidStream.Name = ""
	invalidName := toJSON(invalidStream)

	invalidStream = stream
	invalidStream.Row = ""
	invalidRow := toJSON(invalidStream)

	invalidStream = stream
	invalidStream.Channel = ""
	invalidChannel := toJSON(invalidStream)

	invalidStream = stream
	invalidStream.Host = ""
	invalidHost := toJSON(invalidStream)

	cases := []struct {
		desc        string
		req         string
		contentType string
		auth        string
		status      int
	}{
		{
			desc:        "add valid stream",
			req:         valid,
			contentType: contentType,
			auth:        token,
			status:      http.StatusCreated,
		},
		{
			desc:        "add stream with invalid name",
			req:         invalidName,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add stream with invalid row",
			req:         invalidRow,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add stream with invalid channel",
			req:         invalidChannel,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add stream with invalid host",
			req:         invalidHost,
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add stream with invalid token",
			req:         valid,
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "add stream with empty token",
			req:         valid,
			contentType: contentType,
			auth:        "",
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "add stream with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add stream with empty request",
			req:         "",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add stream without content type",
			req:         valid,
			contentType: "",
			auth:        token,
			status:      http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/streams", ts.URL),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestUpdateStream(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	valid := toJSON(stream)

	invalidStream := stream
	invalidStream.Name = ""
	emptyName := toJSON(invalidStream)

	invalidStream = stream
	invalidStream.Row = ""
	emptyRow := toJSON(invalidStream)

	invalidStream = stream
	invalidStream.Channel = ""
	emptyChannel := toJSON(invalidStream)

	cases := []struct {
		desc        string
		req         string
		contentType string
		auth        string
		name        string
		status      int
	}{
		{
			desc:        "update existing stream",
			req:         valid,
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			status:      http.StatusOK,
		},
		{
			desc:        "update stream with invalid name",
			req:         emptyName,
			contentType: contentType,
			auth:        token,
			name:        "",
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update stream with invalid row",
			req:         emptyRow,
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update stream with invalid channel",
			req:         emptyChannel,
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update stream with invalid token",
			req:         valid,
			contentType: contentType,
			auth:        wrong,
			name:        stream.Name,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "update stream with empty token",
			req:         valid,
			contentType: contentType,
			auth:        "",
			name:        stream.Name,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "update stream with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update stream with empty request",
			req:         "",
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update stream without content type",
			req:         valid,
			contentType: "",
			auth:        token,
			name:        stream.Name,
			status:      http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/streams/%s", ts.URL, tc.name),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestListStreams(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc        string
		contentType string
		auth        string
		status      int
	}{
		{
			desc:        "list streams with valid token",
			contentType: contentType,
			auth:        token,
			status:      http.StatusOK,
		},
		{
			desc:        "list streams with empty token",
			contentType: contentType,
			auth:        "",
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "list streams with wrong token",
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodGet,
			url:         fmt.Sprintf("%s/streams", ts.URL),
			contentType: tc.contentType,
			token:       tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestViewStream(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc        string
		contentType string
		auth        string
		name        string
		status      int
	}{
		{
			desc:        "view stream with valid token",
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			status:      http.StatusOK,
		},
		{
			desc:        "view stream with emtpy name",
			contentType: contentType,
			auth:        token,
			name:        "",
			status:      http.StatusBadRequest,
		},
		{
			desc:        "view stream with empty token",
			contentType: contentType,
			auth:        "",
			name:        stream.Name,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "view stream with wrong token",
			contentType: contentType,
			auth:        wrong,
			name:        stream.Name,
			status:      http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodGet,
			url:         fmt.Sprintf("%s/streams/%s", ts.URL, tc.name),
			contentType: tc.contentType,
			token:       tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestCreateRule(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	ts := newServer(svc)
	defer ts.Close()

	validReq := ruleReq{
		token:          token,
		ID:             "id",
		SQL:            sql,
		Host:           url,
		Port:           "",
		Channel:        channel,
		Subtopic:       "",
		SendMetaToSink: false,
	}

	valid := toJSON(validReq)
	_ = valid

	invalidReq := validReq
	invalidReq.token = ""
	emptyToken := toJSON(invalidReq)

	invalidReq = validReq
	invalidReq.ID = ""
	emptyID := toJSON(invalidReq)

	invalidReq = validReq
	invalidReq.SQL = ""
	emptySQL := toJSON(invalidReq)

	invalidReq = validReq
	invalidReq.Channel = ""
	emptyChannel := toJSON(invalidReq)

	cases := []struct {
		desc        string
		req         string
		contentType string
		auth        string
		status      int
	}{
		{
			desc:        "add rule with required data",
			req:         valid,
			contentType: contentType,
			auth:        token,
			status:      http.StatusCreated,
		},
		{
			desc:        "add rule with wrong token",
			req:         valid,
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "add rule with empty token",
			req:         emptyToken,
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "add rule with empty ID",
			req:         emptyID,
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add rule with empty sql",
			req:         emptySQL,
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add rule with empty channel",
			req:         emptyChannel,
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add rule with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add rule with empty request",
			req:         "",
			contentType: contentType,
			auth:        token,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "add rule without content type",
			req:         valid,
			contentType: "",
			auth:        token,
			status:      http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPost,
			url:         fmt.Sprintf("%s/rules", ts.URL),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestUpdateRule(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token, rule)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	validReq := ruleReq{
		token:          token,
		ID:             rule.ID,
		SQL:            sql,
		Host:           url,
		Port:           "",
		Channel:        channel,
		Subtopic:       "",
		SendMetaToSink: false,
	}

	valid := toJSON(validReq)
	_ = valid

	invalidReq := validReq
	invalidReq.token = ""
	emptyToken := toJSON(invalidReq)

	invalidReq = validReq
	invalidReq.ID = ""
	emptyID := toJSON(invalidReq)

	invalidReq = validReq
	invalidReq.SQL = ""
	emptySQL := toJSON(invalidReq)

	invalidReq = validReq
	invalidReq.Channel = ""
	emptyChannel := toJSON(invalidReq)

	cases := []struct {
		desc        string
		req         string
		contentType string
		auth        string
		id          string
		status      int
	}{
		{
			desc:        "update rule with required data",
			req:         valid,
			contentType: contentType,
			auth:        token,
			id:          validReq.ID,
			status:      http.StatusOK,
		},
		{
			desc:        "update rule with wrong token",
			req:         valid,
			contentType: contentType,
			auth:        wrong,
			id:          validReq.ID,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "update rule with empty token",
			req:         emptyToken,
			contentType: contentType,
			auth:        wrong,
			id:          validReq.ID,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "update rule with empty ID",
			req:         emptyID,
			contentType: contentType,
			auth:        wrong,
			id:          "",
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update rule with empty sql",
			req:         emptySQL,
			contentType: contentType,
			auth:        wrong,
			id:          validReq.ID,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update rule with empty channel",
			req:         emptyChannel,
			contentType: contentType,
			auth:        wrong,
			id:          validReq.ID,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update rule with invalid request format",
			req:         "}",
			contentType: contentType,
			auth:        token,
			id:          validReq.ID,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update rule with empty request",
			req:         "",
			contentType: contentType,
			auth:        token,
			id:          validReq.ID,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "update rule without content type",
			req:         valid,
			contentType: "",
			auth:        token,
			id:          validReq.ID,
			status:      http.StatusUnsupportedMediaType,
		},
	}

	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodPut,
			url:         fmt.Sprintf("%s/rules/%s", ts.URL, tc.id),
			contentType: tc.contentType,
			token:       tc.auth,
			body:        strings.NewReader(tc.req),
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestListRules(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc        string
		contentType string
		auth        string
		status      int
	}{
		{
			desc:        "list rules with valid token",
			contentType: contentType,
			auth:        token,
			status:      http.StatusOK,
		},
		{
			desc:        "list rules with empty token",
			contentType: contentType,
			auth:        "",
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "list rules with empty token",
			contentType: contentType,
			auth:        wrong,
			status:      http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodGet,
			url:         fmt.Sprintf("%s/rules", ts.URL),
			contentType: tc.contentType,
			token:       tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestViewRule(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token, rule)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc        string
		contentType string
		auth        string
		id          string
		status      int
	}{
		{
			desc:        "view rule with valid token",
			contentType: contentType,
			auth:        token,
			id:          rule.ID,
			status:      http.StatusOK,
		},
		{
			desc:        "view rule with emtpy name",
			contentType: contentType,
			auth:        token,
			id:          "",
			status:      http.StatusBadRequest,
		},
		{
			desc:        "view rule with empty token",
			contentType: contentType,
			auth:        "",
			id:          rule.ID,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "view rule with empty token",
			contentType: contentType,
			auth:        wrong,
			id:          rule.ID,
			status:      http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodGet,
			url:         fmt.Sprintf("%s/rules/%s", ts.URL, tc.id),
			contentType: tc.contentType,
			token:       tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestDelete(t *testing.T) {
	svc := mocks.NewService(
		map[string]string{token: email, token2: email2},
		map[string]string{channel: email, channel2: email2},
		url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token, rule)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token2, rule2)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc        string
		contentType string
		auth        string
		name        string
		kuiperType  string
		status      int
	}{
		{
			desc:        "delete existing stream with valid token",
			contentType: contentType,
			auth:        token,
			name:        stream.Name,
			kuiperType:  "streams",
			status:      http.StatusOK,
		},
		{
			desc:        "delete existing rule with valid token",
			contentType: contentType,
			auth:        token,
			name:        rule.ID,
			kuiperType:  "rules",
			status:      http.StatusOK,
		},
		{
			desc:        "delete existing rule with wrong kuiper type",
			contentType: contentType,
			auth:        token2,
			name:        rule2.ID,
			kuiperType:  wrong,
			status:      http.StatusBadRequest,
		},
		{
			desc:        "delete existing rule with empty id",
			contentType: contentType,
			auth:        token2,
			name:        "",
			kuiperType:  wrong,
			status:      http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodDelete,
			url:         fmt.Sprintf("%s/%s/%s", ts.URL, tc.kuiperType, tc.name),
			contentType: tc.contentType,
			token:       tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestRuleStatus(t *testing.T) {
	svc := mocks.NewService(
		map[string]string{token: email, token2: email2},
		map[string]string{channel: email, channel2: email2},
		url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token, rule)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token2, rule2)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc        string
		contentType string
		auth        string
		id          string
		status      int
	}{
		{
			desc:        "rule status with valid token",
			contentType: contentType,
			auth:        token,
			id:          rule.ID,
			status:      http.StatusOK,
		},
		{
			desc:        "rule status with emtpy name",
			contentType: contentType,
			auth:        token,
			id:          "",
			status:      http.StatusBadRequest,
		},
		{
			desc:        "rule status with empty token",
			contentType: contentType,
			auth:        "",
			id:          rule.ID,
			status:      http.StatusUnauthorized,
		},
		{
			desc:        "rule status with wrong token",
			contentType: contentType,
			auth:        wrong,
			id:          rule.ID,
			status:      http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client:      ts.Client(),
			method:      http.MethodGet,
			url:         fmt.Sprintf("%s/rules/%s/status", ts.URL, tc.id),
			contentType: tc.contentType,
			token:       tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

func TestControlRule(t *testing.T) {
	svc := mocks.NewService(map[string]string{token: email}, map[string]string{channel: email}, url)

	_, err := svc.CreateStream(context.Background(), token, stream)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))
	_, err = svc.CreateRule(context.Background(), token, rule)
	require.Nil(t, err, fmt.Sprintf("unexpected error: %s", err))

	ts := newServer(svc)
	defer ts.Close()

	cases := []struct {
		desc   string
		auth   string
		id     string
		action string
		status int
	}{
		{
			desc:   "control rule with valid token",
			auth:   token,
			id:     rule.ID,
			action: action,
			status: http.StatusOK,
		},
		{
			desc:   "control rule with invalid action",
			auth:   token,
			id:     "",
			action: wrong,
			status: http.StatusBadRequest,
		},
		{
			desc:   "control rule with emtpy name",
			auth:   token,
			id:     "",
			action: action,
			status: http.StatusBadRequest,
		},
		{
			desc:   "control rule with empty token",
			auth:   "",
			id:     rule.ID,
			action: action,
			status: http.StatusUnauthorized,
		},
		{
			desc:   "control rule with wrong token",
			auth:   wrong,
			id:     rule.ID,
			action: action,
			status: http.StatusUnauthorized,
		},
	}
	for _, tc := range cases {
		req := testRequest{
			client: ts.Client(),
			method: http.MethodPost,
			url:    fmt.Sprintf("%s/rules/%s/%s", ts.URL, tc.id, tc.action),
			token:  tc.auth,
		}
		res, err := req.make()
		assert.Nil(t, err, fmt.Sprintf("%s: unexpected error %s", tc.desc, err))
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s: expected status code %d got %d", tc.desc, tc.status, res.StatusCode))
	}
}

type ruleReq struct {
	token          string
	ID             string `json:"id"`
	SQL            string `json:"sql"`
	Host           string `json:"host"`
	Port           string `json:"port"`
	Channel        string `json:"channel"`
	Subtopic       string `json:"subtopic"`
	SendMetaToSink bool   `json:"send_meta_to_sink"`
}
