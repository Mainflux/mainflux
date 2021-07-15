// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/rules"
)

var _ rules.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     rules.Service
}

// MetricsMiddleware instruments core service by tracking request count and latency.
func MetricsMiddleware(svc rules.Service, counter metrics.Counter, latency metrics.Histogram) rules.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) Info(ctx context.Context) (info rules.Info, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "info").Add(1)
		ms.latency.With("method", "info").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Info(ctx)
}

func (ms *metricsMiddleware) CreateStream(ctx context.Context, token string, stream rules.Stream) (result string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_stream").Add(1)
		ms.latency.With("method", "create_stream").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.CreateStream(ctx, token, stream)
}

func (ms *metricsMiddleware) UpdateStream(ctx context.Context, token string, stream rules.Stream) (result string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_stream").Add(1)
		ms.latency.With("method", "update_stream").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateStream(ctx, token, stream)
}

func (ms *metricsMiddleware) ListStreams(ctx context.Context, token string) (streams []string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_streams").Add(1)
		ms.latency.With("method", "list_streams").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListStreams(ctx, token)
}

func (ms *metricsMiddleware) ViewStream(ctx context.Context, token, id string) (stream rules.StreamInfo, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_stream").Add(1)
		ms.latency.With("method", "view_stream").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewStream(ctx, token, id)
}

func (ms *metricsMiddleware) Delete(ctx context.Context, token string, id string, kuiperType string) (result string, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "delete_"+kuiperType).Add(1)
		ms.latency.With("method", "delete_"+kuiperType).Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.Delete(ctx, token, id, kuiperType)
}

func (ms *metricsMiddleware) CreateRule(ctx context.Context, token string, rule rules.Rule) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "create_rule").Add(1)
		ms.latency.With("method", "create_rule").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.CreateRule(ctx, token, rule)
}

func (ms *metricsMiddleware) UpdateRule(ctx context.Context, token string, rule rules.Rule) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "update_rule").Add(1)
		ms.latency.With("method", "update_rule").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.UpdateRule(ctx, token, rule)
}

func (ms *metricsMiddleware) ListRules(ctx context.Context, token string) ([]rules.RuleInfo, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "list_rules").Add(1)
		ms.latency.With("method", "list_rules").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListRules(ctx, token)
}

func (ms *metricsMiddleware) ViewRule(ctx context.Context, token, id string) (rule rules.Rule, err error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "view_rule").Add(1)
		ms.latency.With("method", "view_rule").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ViewRule(ctx, token, id)
}

func (ms *metricsMiddleware) RuleStatus(ctx context.Context, token, name string) (map[string]interface{}, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "get_rule_status").Add(1)
		ms.latency.With("method", "get_rule_status").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.RuleStatus(ctx, token, name)
}

func (ms *metricsMiddleware) ControlRule(ctx context.Context, token, name, action string) (string, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "control_rule").Add(1)
		ms.latency.With("method", "control_rule").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ControlRule(ctx, token, name, action)
}
