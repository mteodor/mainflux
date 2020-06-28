// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/mainflux/mainflux/certs"
)

var _ certs.Service = (*metricsMiddleware)(nil)

type metricsMiddleware struct {
	counter metrics.Counter
	latency metrics.Histogram
	svc     certs.Service
}

// MetricsMiddleware instruments core service by tracking request count and
// latency.
func MetricsMiddleware(svc certs.Service, counter metrics.Counter, latency metrics.Histogram) certs.Service {
	return &metricsMiddleware{
		counter: counter,
		latency: latency,
		svc:     svc,
	}
}

func (ms *metricsMiddleware) IssueCert(ctx context.Context, token, thingID string, daysValid string, keyBits int, keyType string) (certs.Cert, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "issue").Add(1)
		ms.latency.With("method", "issue").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.IssueCert(ctx, token, thingID, daysValid, keyBits, keyType)
}

func (ms *metricsMiddleware) ListCertificates(ctx context.Context, token, thingID string, offset, limit uint64) (certs.CertsPage, error) {
	defer func(begin time.Time) {
		ms.counter.With("method", "issue").Add(1)
		ms.latency.With("method", "issue").Observe(time.Since(begin).Seconds())
	}(time.Now())

	return ms.svc.ListCertificates(ctx, token, thingID, offset, limit)
}
