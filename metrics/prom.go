package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type PromMetrics struct {
	buckets  []float64
	registry *prometheus.Registry
}

func NewPromMetrics() *PromMetrics {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	return &PromMetrics{
		registry: registry,
		buckets:  nil,
	}
}

func (m *PromMetrics) MetricsHandler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *PromMetrics) NewJobMetric(queue string) JobMetric {
	reg := m.registry
	jobsTotal := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_processed",
			Help: "Total number of jobs processed",
		}, []string{"queue", "class"},
	)
	jobsOk := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_ok",
			Help: "Number of jobs completed successfully",
		}, []string{"queue", "class"},
	)
	jobsFailed := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_failed",
			Help: "Failed number of jobs",
		}, []string{"queue", "class"},
	)
	return &promJobMetrics{
		jobsTotal:  jobsTotal,
		jobsOk:     jobsOk,
		jobsFailed: jobsFailed,
	}
}

func (m *PromMetrics) NewApiMetric(handlerName string) ApiMetric {
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
	return &promApiMetrics{
		requestsTotal:   requestsTotal,
		requestDuration: requestDuration,
		requestSize:     requestSize,
		responseSize:    responseSize,
	}
}

type promJobMetrics struct {
	jobsTotal  *prometheus.CounterVec
	jobsOk     *prometheus.CounterVec
	jobsFailed *prometheus.CounterVec
}

func (m *promJobMetrics) AddStartedJob(queueName string, jobType string) {
	m.jobsTotal.With(prometheus.Labels{"queue": queueName, "class": jobType}).Add(1)
}

func (m *promJobMetrics) AddCompletedJob(queueName string, jobType string, success bool) {
	if success {
		m.jobsOk.With(prometheus.Labels{"queue": queueName, "class": jobType}).Add(1)
		return
	}
	m.jobsFailed.With(prometheus.Labels{"queue": queueName, "class": jobType}).Add(1)
}

type promApiMetrics struct {
	requestsTotal   *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
	requestSize     *prometheus.SummaryVec
	responseSize    *prometheus.SummaryVec
}

func (m *promApiMetrics) AddResponse(method string, responseCode int, requestSize int64, responseSize int64, responseTime float64) {
	label := prometheus.Labels{"method": method, "code": strconv.Itoa(responseCode)}
	m.requestsTotal.With(label).Add(1)
	m.requestSize.With(label).Observe(float64(requestSize))
	m.requestDuration.With(label).Observe(float64(responseTime))
	m.responseSize.With(label).Observe(float64(responseSize))
}
