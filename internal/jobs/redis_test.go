package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/testutil"
)

func TestRedisJobs(t *testing.T) {
	// redis jobs and cache
	if a, ok := testutil.CheckTestRedisClient(); !ok {
		t.Skip(a)
		return
	}
	client := testutil.MustOpenTestRedisClient()
	rtJobs := NewRedisJobs(client, fmt.Sprintf("queue:%d:%d", os.Getpid(), time.Now().UnixNano()))
	testJobs(t, rtJobs)
}
