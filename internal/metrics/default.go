package metrics

import "net/http"

type DefaultMetric struct{}

func NewDefaultMetric() *DefaultMetric {
	return &DefaultMetric{}
}

func (m *DefaultMetric) NewJobMetric(queue string) JobMetric {
	return &DefaultMetric{}
}

func (m *DefaultMetric) NewApiMetric(handlerName string) ApiMetric {
	return &DefaultMetric{}
}

func (m *DefaultMetric) MetricsHandler() http.Handler {
	return nil
}

func (m *DefaultMetric) AddStartedJob(queueName string, jobType string) {
	return
}

func (m *DefaultMetric) AddCompletedJob(queueName string, jobType string, success bool) {
	return
}

func (m *DefaultMetric) AddResponse(method string, responseCode int, requestSize int64, responseSize int64, responseTime float64) {
	return
}
