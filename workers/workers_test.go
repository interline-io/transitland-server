package workers

import (
	"fmt"
	"testing"

	"github.com/interline-io/transitland-server/config"
)

var (
	feeds = []string{"BA", "SF", "AC", "CT"}
)

func testJobs(t *testing.T, rtJobs config.JobQueue) {
	foundJobs := make([]Job, 0, len(feeds))
	jobfunc := func(job Job) error {
		fmt.Println("got job:", job)
		foundJobs = append(foundJobs, job)
		if len(foundJobs) == len(feeds) {
			rtJobs.Stop()
		}
		return nil
	}
	rtJobs.AddWorker(jobfunc, 1)
	for _, feed := range feeds {
		url := fmt.Sprintf("test/%s.pb", feed)
		rtJobs.AddJob(Job{JobType: "test", Args: []string{feed, url}})
	}
	rtJobs.Run()
	if len(foundJobs) != len(feeds) {
		t.Errorf("got %d jobs, expected %d", len(foundJobs), len(feeds))
	}
}
