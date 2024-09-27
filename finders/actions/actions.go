package actions

import (
	"context"
	"io"

	"github.com/interline-io/transitland-server/model"
)

func init() {
	var _ model.Actions = &Actions{}
}

type Actions struct{}

func (Actions) StaticFetch(ctx context.Context, feedId string, feedSrc io.Reader, feedUrl string) (*model.FeedVersionFetchResult, error) {
	return StaticFetch(ctx, feedId, feedSrc, feedUrl)
}

func (Actions) RTFetch(ctx context.Context, target string, feedId string, feedUrl string, urlType string) error {
	return RTFetch(ctx, target, feedId, feedUrl, urlType)
}

func (Actions) ValidateUpload(ctx context.Context, src io.Reader, feedURL *string, rturls []string) (*model.ValidationReport, error) {
	return ValidateUpload(ctx, src, feedURL, rturls)
}

func (Actions) GBFSFetch(ctx context.Context, feedId string, feedUrl string) error {
	return GBFSFetch(ctx, feedId, feedUrl)
}

func (Actions) FetchEnqueue(ctx context.Context, feedIds []string, urlTypes []string, ignoreFetchWait bool) error {
	return FetchEnqueue(ctx, feedIds, urlTypes, ignoreFetchWait)
}
