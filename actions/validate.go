package actions

import (
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tlcsv"
	"github.com/interline-io/transitland-lib/validator"
	"github.com/interline-io/transitland-server/model"
)

// ValidateUpload takes a file Reader and produces a validation package containing errors, warnings, file infos, service levels, etc.
func ValidateUpload(ctx context.Context, src io.Reader, feedURL *string, rturls []string) (*model.ValidationReport, error) {
	cfg := model.ForContext(ctx)

	// Check inputs
	rturlsok := []string{}
	for _, rturl := range rturls {
		if checkurl(rturl) {
			rturlsok = append(rturlsok, rturl)
		}
	}
	rturls = rturlsok
	if len(rturls) > 3 {
		rturls = rturls[0:3]
	}
	if feedURL == nil || !checkurl(*feedURL) {
		feedURL = nil
	}
	//////
	result := model.ValidationReport{}
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
			result.FailureReason = strptr("Could not read file")
			return &result, nil
		}
	} else if feedURL != nil {
		var err error
		reader, err = tlcsv.NewReader(*feedURL)
		if err != nil {
			result.FailureReason = strptr("Could not load URL")
			return &result, nil
		}
	} else {
		result.FailureReason = strptr("No feed specified")
		return &result, nil
	}

	if err := reader.Open(); err != nil {
		result.FailureReason = strptr("Could not read file")
		return &result, nil
	}

	// Perform validation
	opts := validator.Options{
		BestPractices:            true,
		CheckFileLimits:          true,
		IncludeServiceLevels:     true,
		IncludeRouteGeometries:   true,
		IncludeEntities:          true,
		IncludeRealtimeJson:      true,
		IncludeEntitiesLimit:     10_000,
		MaxRTMessageSize:         10_000_000,
		ValidateRealtimeMessages: rturls,
	}
	if cfg.ValidateLargeFiles {
		opts.CheckFileLimits = false
	}
	opts.ErrorLimit = 1_000

	vt, err := validator.NewValidator(reader, opts)
	if err != nil {
		result.FailureReason = strptr("Could not validate file")
		return &result, nil
	}
	r, err := vt.Validate()
	if err != nil {
		result.FailureReason = strptr("Could not validate file")
		return &result, nil
	}

	// Some mapping is necessary because most gql models have some extra fields not in the base tl models.
	result.Success = r.Success.Val
	if r.FailureReason.Valid {
		result.FailureReason = &r.FailureReason.Val
	}
	result.Details = &model.ValidationReportDetails{}
	result.Details.Sha1 = r.Details.SHA1.Val
	result.Details.EarliestCalendarDate = &r.Details.EarliestCalendarDate
	result.Details.LatestCalendarDate = &r.Details.LatestCalendarDate
	for _, eg := range r.Errors {
		if eg == nil {
			continue
		}
		eg2 := model.ValidationReportErrorGroup{
			Filename:  eg.Filename,
			Field:     eg.Field,
			ErrorCode: eg.ErrorCode,
			ErrorType: eg.ErrorType,
			GroupKey:  eg.GroupKey,
			Count:     eg.Count,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationReportError{
				Filename:  eg.Filename,
				Field:     eg.Field,
				ErrorType: eg.ErrorType,
				ErrorCode: eg.ErrorCode,
				GroupKey:  eg.GroupKey,
				Line:      err.Line,
				EntityID:  err.EntityID,
				Message:   err.Message,
				Geometry:  &err.Geometry,
			}
			eg2.Errors = append(eg2.Errors, &err2)
		}
		result.Errors = append(result.Errors, &eg2)
	}
	for _, eg := range r.Warnings {
		if eg == nil {
			continue
		}
		eg2 := model.ValidationReportErrorGroup{
			Filename:  eg.Filename,
			Field:     eg.Field,
			ErrorCode: eg.ErrorCode,
			ErrorType: eg.ErrorType,
			GroupKey:  eg.GroupKey,
			Count:     eg.Count,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationReportError{
				Filename:  eg.Filename,
				Field:     eg.Field,
				ErrorType: eg.ErrorType,
				ErrorCode: eg.ErrorCode,
				GroupKey:  eg.GroupKey,
				Line:      err.Line,
				EntityID:  err.EntityID,
				Message:   err.Message,
				Geometry:  &err.Geometry,
			}
			eg2.Errors = append(eg2.Errors, &err2)
		}
		result.Warnings = append(result.Warnings, &eg2)
	}
	for _, v := range r.Details.FeedInfos {
		result.Details.FeedInfos = append(result.Details.FeedInfos, &model.FeedInfo{FeedInfo: v})
	}
	for _, v := range r.Details.Files {
		result.Details.Files = append(result.Details.Files, &model.FeedVersionFileInfo{FeedVersionFileInfo: v})
	}
	for _, v := range r.Details.ServiceLevels {
		result.Details.ServiceLevels = append(result.Details.ServiceLevels, &model.FeedVersionServiceLevel{FeedVersionServiceLevel: v})
	}
	for _, v := range r.Details.Agencies {
		result.Details.Agencies = append(result.Details.Agencies, &model.Agency{Agency: v})
	}
	for _, v := range r.Details.Routes {
		result.Details.Routes = append(result.Details.Routes, &model.Route{Route: v})
	}
	for _, v := range r.Details.Stops {
		result.Details.Stops = append(result.Details.Stops, &model.Stop{Stop: v})
	}
	for _, v := range r.Details.Realtime {
		result.Details.Realtime = append(result.Details.Realtime, &model.ValidationRealtimeResult{
			URL:  v.Url,
			JSON: tt.NewMap(v.Json),
		})
	}
	return &result, nil
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

func strptr(v string) *string {
	return &v
}
