package workers

import (
	"context"
	"fmt"

	"github.com/interline-io/transitland-server/rtcache"
)

//

type RTFetchWorker struct{}

func (w *RTFetchWorker) Run(ctx context.Context, opts JobOptions, job rtcache.Job) error {
	fmt.Printf("fetch worker! job: %#v\n", job)
	return nil
}
