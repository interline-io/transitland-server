// Copyright 2022 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// FROM:

// Adapted from
// https://github.com/prometheus/client_golang/tree/main/examples/middleware/httpmiddleware
// and
// https://github.com/brancz/prometheus-example-app/blob/master/main.go
package metrics

import (
	"net/http"

	"github.com/interline-io/transitland-lib/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpWrapper interface {
	// WrapHandler wraps the given HTTP handler for instrumentation.
	WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc
}

type httpMiddleware struct {
	buckets  []float64
	registry prometheus.Registerer
}

// WrapHandler wraps the given HTTP handler for instrumentation:
// It registers four metric collectors (if not already done) and reports HTTP
// metrics to the (newly or already) registered collectors.
// Each has a constant label named "handler" with the provided handlerName as
// value.
func (m *httpMiddleware) WrapHandler(handlerName string, handler http.Handler) http.HandlerFunc {
	log.Infof("Registering metrics middleware: %s", handlerName)
	reg := prometheus.WrapRegistererWith(prometheus.Labels{"handler": handlerName}, m.registry)
	requestsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Tracks the number of HTTP requests.",
		}, []string{"method", "code"},
	)
	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Tracks the latencies for HTTP requests.",
			Buckets: m.buckets,
		},
		[]string{"method", "code"},
	)
	requestSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Tracks the size of HTTP requests.",
		},
		[]string{"method", "code"},
	)
	responseSize := promauto.With(reg).NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Tracks the size of HTTP responses.",
		},
		[]string{"method", "code"},
	)

	// Wraps the provided http.Handler to observe the request result with the provided metrics.
	base := promhttp.InstrumentHandlerCounter(
		requestsTotal,
		promhttp.InstrumentHandlerDuration(
			requestDuration,
			promhttp.InstrumentHandlerRequestSize(
				requestSize,
				promhttp.InstrumentHandlerResponseSize(
					responseSize,
					http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
						handler.ServeHTTP(writer, r)
					}),
				),
			),
		),
	)

	return base.ServeHTTP
}

// New returns a Middleware interface.
func NewHttpMiddleware(registry prometheus.Registerer, buckets []float64) HttpWrapper {
	if buckets == nil {
		buckets = prometheus.ExponentialBuckets(0.1, 1.5, 5)
	}
	return &httpMiddleware{
		buckets:  buckets,
		registry: registry,
	}
}

type passthroughHttpMiddleware struct{}

func (m *passthroughHttpMiddleware) WrapHandler(handlerName string, h http.Handler) http.HandlerFunc {
	return h.ServeHTTP
}

func NewPassthroughMiddleware() HttpWrapper {
	return &passthroughHttpMiddleware{}
}
