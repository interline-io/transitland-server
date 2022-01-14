package workers

import (
	"context"
	"errors"
	"fmt"

	"github.com/interline-io/transitland-lib/rt"
	"github.com/interline-io/transitland-server/rtcache"
	"google.golang.org/protobuf/proto"
)

//

type RTFetchWorker struct{}

func (w *RTFetchWorker) Run(ctx context.Context, opts JobOptions, job rtcache.Job) error {
	fmt.Printf("fetch worker! job: %#v\n", job)
	if len(job.Args) != 2 {
		return errors.New("feed and url required")
	}
	feed := job.Args[0]
	url := job.Args[1]
	msg, err := rt.ReadURL(url)
	if err != nil {
		return err
	}
	fmt.Println("got msg:", msg)
	rtdata, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return opts.cache.AddData(feed, rtdata)
}
