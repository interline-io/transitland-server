package monitoring

import (
	"sync"
	"time"
)

// QueryStats tracks query execution patterns
type QueryStats struct {
	mu           sync.RWMutex
	queryMetrics map[string]*QueryMetric
}

type QueryMetric struct {
	SQL            string
	ExecutionCount int64
	TotalExecTime  time.Duration
	TotalPlanTime  time.Duration
	LastExecuted   time.Time
}

// NewQueryStats creates a new query statistics tracker
func NewQueryStats() *QueryStats {
	return &QueryStats{
		queryMetrics: make(map[string]*QueryMetric),
	}
}

// RecordQuery records execution statistics for a query
func (qs *QueryStats) RecordQuery(sql string, execTime, planTime time.Duration) {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	metric, exists := qs.queryMetrics[sql]
	if !exists {
		metric = &QueryMetric{SQL: sql}
		qs.queryMetrics[sql] = metric
	}

	metric.ExecutionCount++
	metric.TotalExecTime += execTime
	metric.TotalPlanTime += planTime
	metric.LastExecuted = time.Now()
}

// GetStats returns current query statistics
func (qs *QueryStats) GetStats() map[string]*QueryMetric {
	qs.mu.RLock()
	defer qs.mu.RUnlock()

	result := make(map[string]*QueryMetric)
	for k, v := range qs.queryMetrics {
		// Create a copy to avoid race conditions
		result[k] = &QueryMetric{
			SQL:            v.SQL,
			ExecutionCount: v.ExecutionCount,
			TotalExecTime:  v.TotalExecTime,
			TotalPlanTime:  v.TotalPlanTime,
			LastExecuted:   v.LastExecuted,
		}
	}
	return result
}
