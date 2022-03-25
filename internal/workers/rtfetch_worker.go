package workers

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/request"
	"github.com/interline-io/transitland-server/internal/jobs"
	"github.com/interline-io/transitland-server/model"
	"google.golang.org/protobuf/proto"
)

type RTFetchWorker struct {
	Target       string `json:"target"`
	Url          string `json:"url"`
	SourceType   string `json:"source_type"`
	SourceFeedID string `json:"source_feed_id"`
}

func (w *RTFetchWorker) Run(ctx context.Context, job jobs.Job) error {
	log := job.Opts.Logger.With().Str("target", w.Target).Str("source_feed_id", w.SourceFeedID).Str("source_type", w.SourceType).Str("url", w.Url).Logger()
	// Find feed
	rtfeeds, err := job.Opts.Finder.FindFeeds(nil, nil, nil, &model.FeedFilter{OnestopID: &w.SourceFeedID})
	if err != nil {
		log.Error().Err(err).Msg("rtfetch worker: error loading source feed")
		return err
	}
	if len(rtfeeds) == 0 {
		log.Error().Err(err).Msg("rtfetch worker: source feed not found")
		return errors.New("feed not found")
	}
	rtfeed := rtfeeds[0]
	// Prepare auth
	// Note: Only HTTP(S) allowed; AllowFTP/AllowS3/AllowLocal options not passed in.
	var reqOpts []request.RequestOption
	if rtfeed.Authorization.Type != "" {
		secret := tl.Secret{}
		var err error
		secret, err = rtfeed.MatchSecrets(job.Opts.Secrets)
		if err != nil {
			log.Error().Err(err).Msg("rtfetch worker: secret match failed")
			return err
		}
		reqOpts = append(reqOpts, request.WithAuth(secret, rtfeed.Authorization))
	}
	// Make request
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req := request.NewRequest(w.Url, reqOpts...)
	reqBody, err := req.Request(ctx)
	if err != nil {
		log.Error().Err(err).Msg("rtfetch worker: request failed")
		return err
	}
	defer reqBody.Close()
	// Test this is valid protobuf
	rtdata, _ := ioutil.ReadAll(reqBody)
	rtmsg := pb.FeedMessage{}
	if err := proto.Unmarshal(rtdata, &rtmsg); err != nil {
		log.Error().Err(err).Msg("rtfetch worker: failed to parse response")
		return err
	}
	// Save to cache
	key := fmt.Sprintf("rtdata:%s:%s", w.Target, w.SourceType)
	log.Info().Int("bytes", len(rtdata)).Str("url", w.Url).Msg("rtfetch worker: success")
	return job.Opts.RTFinder.AddData(key, rtdata)
}
