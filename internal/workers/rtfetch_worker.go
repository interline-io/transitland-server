package workers

import (
	"context"
	"errors"
	"fmt"

	"github.com/interline-io/transitland-lib/rt"
	"github.com/interline-io/transitland-server/internal/jobs"
	"google.golang.org/protobuf/proto"
)

type RTFetchWorker struct{}

func (w *RTFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	// fmt.Printf("fetch worker! job: %#v\n", job)
	if len(job.Args) != 3 {
		return errors.New("feed, type and url required")
	}
	feed := job.Args[0]
	ftype := job.Args[1]
	url := job.Args[2]
	var rtdata []byte
	msg, err := rt.ReadURL(url)
	if err != nil {
		return err
	}
	rtdata, err = proto.Marshal(msg)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("rtdata:%s:%s", feed, ftype)
	return job.Opts.RTFinder.AddData(key, rtdata)
}