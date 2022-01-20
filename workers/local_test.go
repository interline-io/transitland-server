package workers

import (
	"testing"
)

func TestLocalJobs(t *testing.T) {
	rtJobs := NewLocalJobs()
	testJobs(t, rtJobs)
}
