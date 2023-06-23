package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/dbutil"
)

func TestRedisJobs(t *testing.T) {
	// redis jobs and cache
	if a, ok := dbutil.CheckTestRedisClient(); !ok {
		t.Skip(a)
		return
	}
	client := dbutil.MustOpenTestRedisClient()
	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
	testJobs(t, rtJobs)
}
