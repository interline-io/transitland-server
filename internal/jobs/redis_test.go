package jobs

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestRedisJobs(t *testing.T) {
	// redis jobs and cache
	if a, ok := testutil.CheckTestRedisClient(); !ok {
		t.Skip(a)
		return
	}
	client := testutil.MustOpenTestRedisClient()

	t.Run("simple", func(t *testing.T) {
		// Ugly :(
		rtJobs := NewRedisJobs(client, fmt.Sprintf("queue1:%d:%d", os.Getpid(), time.Now().UnixNano()))
		count := int64(0)
		testGetWorker := func(job Job) (JobWorker, error) {
			w := testWorker{count: &count}
			return &w, nil
		}
		rtJobs.AddWorker("", testGetWorker, JobOptions{}, 1)
		for _, feed := range feeds {
			rtJobs.AddJob(Job{JobType: "test", Unique: false, JobArgs: JobArgs{"feed_id": feed}})
		}
		go func() {
			time.Sleep(200 * time.Millisecond)
			rtJobs.Stop()
		}()
		rtJobs.Run()
		assert.Equal(t, len(feeds), int(count))
	})
	t.Run("unique", func(t *testing.T) {
		// Abuse the job queue
		rtJobs := NewRedisJobs(client, fmt.Sprintf("queue2:%d:%d", os.Getpid(), time.Now().UnixNano()))
		count := int64(0)
		testGetWorker := func(job Job) (JobWorker, error) {
			w := testWorker{count: &count}
			return &w, nil
		}
		for i := 0; i < 10; i++ {
			// 1 job: j=0
			for j := 0; j < 10; j++ {
				job := Job{JobType: "testUnique", Unique: true, JobArgs: JobArgs{"test": fmt.Sprintf("n:%d", j/10)}}
				rtJobs.AddJob(job)
			}
			// 3 jobs; j=3, j=6, j=9... j=0 is not unique
			for j := 0; j < 10; j++ {
				job := Job{JobType: "testUnique", Unique: true, JobArgs: JobArgs{"test": fmt.Sprintf("n:%d", j/3)}}
				rtJobs.AddJob(job)
			}
			// 10 jobs: j=0, j=0, j=2, j=2, j=4, j=4, j=6 j=6, j=8, j=8
			for j := 0; j < 10; j++ {
				job := Job{JobType: "testNotUnique", Unique: false, JobArgs: JobArgs{"test": fmt.Sprintf("n:%d", j/2)}}
				rtJobs.AddJob(job)
			}
		}
		rtJobs.AddWorker("", testGetWorker, JobOptions{}, 4)
		go func() {
			time.Sleep(1000 * time.Millisecond)
			rtJobs.Stop()
		}()
		rtJobs.Run()
		assert.Equal(t, int64(104), count)
	})
	t.Run("deadline", func(t *testing.T) {
		rtJobs := NewRedisJobs(client, fmt.Sprintf("queue3:%d:%d", os.Getpid(), time.Now().UnixNano()))
		count := int64(0)
		testGetWorker := func(job Job) (JobWorker, error) {
			w := testWorker{count: &count}
			return &w, nil
		}
		rtJobs.AddJob(Job{JobType: "testUnique", Unique: false, JobArgs: JobArgs{"test": "test"}, JobDeadline: 0})
		rtJobs.AddJob(Job{JobType: "testUnique", Unique: false, JobArgs: JobArgs{"test": "test"}, JobDeadline: time.Now().Add(1 * time.Hour).Unix()})
		rtJobs.AddJob(Job{JobType: "testUnique", Unique: false, JobArgs: JobArgs{"test": "test"}, JobDeadline: time.Now().Add(-1 * time.Hour).Unix()})
		rtJobs.AddWorker("", testGetWorker, JobOptions{}, 1)
		go func() {
			time.Sleep(100 * time.Millisecond)
			rtJobs.Stop()
		}()
		rtJobs.Run()
		assert.Equal(t, int64(2), count)
	})
}
