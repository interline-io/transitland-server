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
	log := job.Opts.Logger
	if len(job.Args) != 4 {
		return errors.New("feed, type and url required")
	}
	feed := job.Args[0]
	ftype := job.Args[1]
	url := job.Args[2]
	source := job.Args[3]
	var rtdata []byte
	msg, err := rt.ReadURL(url)
	if err != nil {
		log.Error().Err(err).Str("feed_id", feed).Str("source", source).Str("source_type", ftype).Str("url", url).Msg("fetch worker: request failed")
		return err
	}
	rtdata, err = proto.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Str("feed_id", feed).Str("source", source).Str("source_type", ftype).Str("url", url).Msg("fetch worker: failed to parse response")
		return err
	}
	key := fmt.Sprintf("rtdata:%s:%s", feed, ftype)
	return job.Opts.RTFinder.AddData(key, rtdata)
}
