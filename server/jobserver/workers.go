package jobserver

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-server/model"
)

// Default worker definitions

// StaticFetchWorker
type StaticFetchWorker struct {
	FeedUrl    string `json:"feed_url"`
	FeedID     string `json:"feed_id"`
	FetchEpoch int64  `json:"fetch_epoch"`
}

func (w *StaticFetchWorker) Kind() string {
	return "static-fetch"
}

func (w *StaticFetchWorker) Run(ctx context.Context) error {
	result, err := model.ForContext(ctx).Actions.StaticFetch(ctx, w.FeedID, nil, w.FeedUrl)
	if err != nil {
		return err
	} else if result.FetchError != nil {
		return errors.New(*result.FetchError)
	}
	return nil
}

// RTFetchWorker
type RTFetchWorker struct {
	Target       string `json:"target"`
	Url          string `json:"url"`
	SourceType   string `json:"source_type"`
	SourceFeedID string `json:"source_feed_id"`
	FetchEpoch   int64  `json:"fetch_epoch"`
}

func (w *RTFetchWorker) Kind() string {
	return "rt-fetch"
}

func (w *RTFetchWorker) Run(ctx context.Context) error {
	return model.ForContext(ctx).Actions.RTFetch(ctx, w.Target, w.SourceFeedID, w.Url, w.SourceType)
}

// GbfsFetchWorker
type GbfsFetchWorker struct {
	Url        string `json:"url"`
	FeedID     string `json:"feed_id"`
	FetchEpoch int64  `json:"fetch_epoch"`
}

func (w *GbfsFetchWorker) Kind() string {
	return "gbfs-fetch"
}

func (w *GbfsFetchWorker) Run(ctx context.Context) error {
	return model.ForContext(ctx).Actions.GBFSFetch(ctx, w.FeedID, w.Url)
}

// FetchEnqueueWorker
type FetchEnqueueWorker struct {
	IgnoreFetchWait bool     `json:"ignore_fetch_wait"`
	URLTypes        []string `json:"url_types"`
	FeedIDs         []string `json:"feed_ids"`
}

func (w *FetchEnqueueWorker) Kind() string {
	return "fetch-enqueue"
}

func (w *FetchEnqueueWorker) Run(ctx context.Context) error {
	return model.ForContext(ctx).Actions.FetchEnqueue(ctx, w.FeedIDs, w.URLTypes, w.IgnoreFetchWait)
}

// FeedVersionImportWorker
type FeedVersionImportWorker struct {
	FeedVersionID int `json:"feed_version_id"`
}

func (w *FeedVersionImportWorker) Kind() string {
	return "fetch-version-import"
}

func (w *FeedVersionImportWorker) Run(ctx context.Context) error {
	_, err := model.ForContext(ctx).Actions.FeedVersionImport(ctx, w.FeedVersionID)
	return err
}

// FeedVersionUnimportWorker
type FeedVersionUnimportWorker struct {
	FeedVersionID int `json:"feed_version_id"`
}

func (w *FeedVersionUnimportWorker) Kind() string {
	return "fetch-version-unimport"
}

func (w *FeedVersionUnimportWorker) Run(ctx context.Context) error {
	_, err := model.ForContext(ctx).Actions.FeedVersionUnimport(ctx, w.FeedVersionID)
	return err
}

// FeedVersionDeleteWorker
type FeedVersionDeleteWorker struct {
	FeedVersionID int `json:"feed_version_id"`
}

func (w *FeedVersionDeleteWorker) Kind() string {
	return "fetch-version-delete"
}

func (w *FeedVersionDeleteWorker) Run(ctx context.Context) error {
	_, err := model.ForContext(ctx).Actions.FeedVersionDelete(ctx, w.FeedVersionID)
	return err
}
