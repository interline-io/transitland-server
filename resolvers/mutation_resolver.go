package resolvers

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/interline-io/transitland-lib/log"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/causes"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tlcsv"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-lib/validator"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
)

// mutation root

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ValidateGtfs(ctx context.Context, file *graphql.Upload, url *string, rturls []string) (*model.ValidationResult, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return ValidateUpload(r.cfg, src, url, rturls, auth.ForContext(ctx))
}

func (r *mutationResolver) FeedVersionFetch(ctx context.Context, file *graphql.Upload, url *string, feed string) (*model.FeedVersionFetchResult, error) {
	var src io.Reader
	if file != nil {
		src = file.File
	}
	return Fetch(r.cfg, r.finder, src, url, feed, auth.ForContext(ctx))
}

func (r *mutationResolver) FeedVersionImport(ctx context.Context, sha1 string) (*model.FeedVersionImportResult, error) {
	return nil, errors.New("temporarily unavailable")
}

func (r *mutationResolver) FeedVersionUpdate(ctx context.Context, id int, values model.FeedVersionSetInput) (*model.FeedVersion, error) {
	return nil, errors.New("temporarily unavailable")
}

func (r *mutationResolver) FeedVersionUnimport(ctx context.Context, id int) (*model.FeedVersionUnimportResult, error) {
	return nil, errors.New("temporarily unavailable")
}

func (r *mutationResolver) FeedVersionDelete(ctx context.Context, id int) (*model.FeedVersionDeleteResult, error) {
	return nil, errors.New("temporarily unavailable")
}

// Fetch adds a feed version to the database.
func Fetch(cfg config.Config, finder model.Finder, src io.Reader, feedURL *string, feedId string, user *auth.User) (*model.FeedVersionFetchResult, error) {
	if user == nil {
		return nil, errors.New("no user")
	}
	// Find feed
	// feeds, err := cfg.Finder.FindFeeds(nil, nil, nil, &model.FeedFilter{OnestopID: &feed})
	var feeds []tl.Feed
	atx := tldb.NewPostgresAdapterFromDBX(finder.DBX())
	err := atx.Select(&feeds, "select * from current_feeds where onestop_id = ?", feedId)
	if err != nil {
		log.Error().Err(err).Msg("fetch mutation: error loading source feed")
		return nil, err
	}
	if len(feeds) == 0 {
		log.Error().Err(err).Msg("fetch mutation: source feed not found")
		return nil, errors.New("feed not found")
	}
	feed := feeds[0]
	// Prepare request
	opts := fetch.Options{
		URLType:   "manual",
		FetchedAt: time.Now(),
		FeedID:    feed.ID,
		Storage:   cfg.Storage,
		CreatedBy: tt.NewString(user.Name),
	}
	if src != nil {
		// Prepare reader
		tmpfile, err := ioutil.TempFile("", "validator-upload")
		if err != nil {
			// This should result in a failed request
			return nil, err
		}
		io.Copy(tmpfile, src)
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())
		opts.FeedURL = tmpfile.Name()
	} else if feedURL != nil {
		opts.FeedURL = *feedURL
	}
	// Make request
	fv, fr, err := fetch.StaticFetch(atx, opts)
	if err != nil {
		return nil, err
	}
	mr := model.FeedVersionFetchResult{
		FoundSHA1: fr.Found,
	}
	if fr.FetchError == nil {
		mr.FeedVersion = &model.FeedVersion{FeedVersion: fv}
		mr.FetchError = nil
	} else {
		return nil, fr.FetchError
	}
	return &mr, nil
}

type hasContext interface {
	Context() *causes.Context
}

func checkurl(address string) bool {
	if address == "" {
		return false
	}
	u, err := url.Parse(address)
	if err != nil {
		return false
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return true
	}
	return false
}

// ValidateUpload takes a file Reader and produces a validation package containing errors, warnings, file infos, service levels, etc.
func ValidateUpload(cfg config.Config, src io.Reader, feedURL *string, rturls []string, user *auth.User) (*model.ValidationResult, error) {
	// Check inputs
	rturlsok := []string{}
	for _, rturl := range rturls {
		if checkurl(rturl) {
			rturlsok = append(rturlsok, rturl)
		}
	}
	rturls = rturlsok
	if feedURL == nil || !checkurl(*feedURL) {
		feedURL = nil
	}
	//////
	result := model.ValidationResult{}
	result.EarliestCalendarDate = tl.Date{}
	result.LatestCalendarDate = tl.Date{}
	var reader tl.Reader
	if src != nil {
		// Prepare reader
		var err error
		tmpfile, err := ioutil.TempFile("", "validator-upload")
		if err != nil {
			// This should result in a failed request
			return nil, err
		}
		io.Copy(tmpfile, src)
		tmpfile.Close()
		defer os.Remove(tmpfile.Name())
		reader, err = tlcsv.NewReader(tmpfile.Name())
		if err != nil {
			result.FailureReason = "Could not read file"
			return &result, nil
		}
	} else if feedURL != nil {
		var err error
		reader, err = tlcsv.NewReader(*feedURL)
		if err != nil {
			result.FailureReason = "Could not load URL"
			return &result, nil
		}
	} else {
		result.FailureReason = "No feed specified"
		return &result, nil
	}

	if err := reader.Open(); err != nil {
		result.FailureReason = "Could not read file"
		return &result, nil
	}

	// Perform validation
	opts := validator.Options{
		BestPractices:            true,
		CheckFileLimits:          true,
		IncludeServiceLevels:     true,
		IncludeRouteGeometries:   true,
		IncludeEntities:          true,
		IncludeEntitiesLimit:     10000,
		ValidateRealtimeMessages: rturls,
	}
	if cfg.ValidateLargeFiles {
		opts.CheckFileLimits = false
	}

	checker, err := validator.NewValidator(reader, opts)
	if err != nil {
		result.FailureReason = "Could not validate file"
		return &result, nil
	}
	r, err := checker.Validate()
	if err != nil {
		result.FailureReason = "Could not validate file"
		return &result, nil
	}

	// Some mapping is necessary because most gql models have some extra fields not in the base tl models.
	result.Success = r.Success
	result.FailureReason = r.FailureReason
	result.Sha1 = r.SHA1
	result.EarliestCalendarDate = r.EarliestCalendarDate
	result.LatestCalendarDate = r.LatestCalendarDate
	for _, eg := range r.Errors {
		if eg == nil {
			continue
		}
		eg2 := model.ValidationResultErrorGroup{
			Filename:  eg.Filename,
			ErrorType: eg.ErrorType,
			Count:     eg.Count,
			Limit:     eg.Limit,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationResultError{
				Filename: eg.Filename,
				Message:  err.Error(),
			}
			if v, ok := err.(hasContext); ok {
				c := v.Context()
				err2.EntityID = c.EntityID
				err2.Field = c.Field
			}
			eg2.Errors = append(eg2.Errors, &err2)
		}
		result.Errors = append(result.Errors, eg2)
	}
	for _, eg := range r.Warnings {
		if eg == nil {
			continue
		}
		eg2 := model.ValidationResultErrorGroup{
			Filename:  eg.Filename,
			ErrorType: eg.ErrorType,
			Count:     eg.Count,
			Limit:     eg.Limit,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationResultError{
				Filename: eg.Filename,
				Message:  err.Error(),
			}
			if v, ok := err.(hasContext); ok {
				c := v.Context()
				err2.EntityID = c.EntityID
				err2.Field = c.Field
			}
			eg2.Errors = append(eg2.Errors, &err2)
		}
		result.Warnings = append(result.Warnings, eg2)
	}
	for _, v := range r.FeedInfos {
		result.FeedInfos = append(result.FeedInfos, model.FeedInfo{FeedInfo: v})
	}
	for _, v := range r.Files {
		result.Files = append(result.Files, model.FeedVersionFileInfo{FeedVersionFileInfo: v})
	}
	for _, v := range r.ServiceLevels {
		result.ServiceLevels = append(result.ServiceLevels, model.FeedVersionServiceLevel{FeedVersionServiceLevel: v})
	}
	for _, v := range r.Agencies {
		result.Agencies = append(result.Agencies, model.Agency{Agency: v})
	}
	for _, v := range r.Routes {
		result.Routes = append(result.Routes, model.Route{Geometry: v.Geometry, Route: v})
	}
	for _, v := range r.Stops {
		result.Stops = append(result.Stops, model.Stop{Stop: v})
	}
	return &result, nil
}
