package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tlcsv"
	"github.com/interline-io/transitland-lib/validator"
	"github.com/interline-io/transitland-server/model"
)

// ValidateUpload takes a file Reader and produces a validation package containing errors, warnings, file infos, service levels, etc.
func ValidateUpload(ctx context.Context, src io.Reader, feedURL *string, rturls []string) (*model.ValidationResult, error) {
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
	result := model.ValidationResult{}
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
		IncludeRealtimeJson:      true,
		IncludeEntitiesLimit:     10_000,
		MaxRTMessageSize:         10_000_000,
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
	fmt.Println("WARN????", r.Warnings)

	// Some mapping is necessary because most gql models have some extra fields not in the base tl models.
	// result.RawResult = r
	result.Success = r.Success
	result.FailureReason = r.FailureReason
	result.Details = model.ValidationResultDetails{}
	result.Details.Sha1 = r.Details.SHA1
	result.Details.EarliestCalendarDate = r.Details.EarliestCalendarDate
	result.Details.LatestCalendarDate = r.Details.LatestCalendarDate
	for _, eg := range r.Errors {
		if eg == nil {
			continue
		}
		eg2 := model.ValidationResultErrorGroup{
			Filename:  eg.Filename,
			Field:     eg.Field,
			ErrorCode: eg.ErrorCode,
			ErrorType: eg.ErrorType,
			Count:     eg.Count,
			Limit:     eg.Limit,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationResultError{
				Filename:  eg.Filename,
				Field:     eg.Field,
				ErrorType: eg.ErrorType,
				ErrorCode: eg.ErrorCode,
				Line:      err.Line,
				EntityID:  err.EntityID,
				Message:   err.Error(),
				Geometry:  err.Geometry,
			}
			eg2.Errors = append(eg2.Errors, &err2)
		}
		result.Errors = append(result.Errors, eg2)
	}
	jj, _ := json.Marshal(r)
	fmt.Println(string(jj))
	for _, eg := range r.Warnings {
		fmt.Println("WARN2:", eg)
		if eg == nil {
			continue
		}
		eg2 := model.ValidationResultErrorGroup{
			Filename:  eg.Filename,
			Field:     eg.Field,
			ErrorCode: eg.ErrorCode,
			ErrorType: eg.ErrorType,
			Count:     eg.Count,
			Limit:     eg.Limit,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationResultError{
				Filename:  eg.Filename,
				Field:     eg.Field,
				ErrorType: eg.ErrorType,
				ErrorCode: eg.ErrorCode,
				Line:      err.Line,
				EntityID:  err.EntityID,
				Message:   err.Error(),
				Geometry:  err.Geometry,
			}
			eg2.Errors = append(eg2.Errors, &err2)
		}
		result.Warnings = append(result.Warnings, eg2)
	}
	for _, v := range r.Details.FeedInfos {
		result.Details.FeedInfos = append(result.Details.FeedInfos, model.FeedInfo{FeedInfo: v})
	}
	for _, v := range r.Details.Files {
		result.Details.Files = append(result.Details.Files, model.FeedVersionFileInfo{FeedVersionFileInfo: v})
	}
	for _, v := range r.Details.ServiceLevels {
		result.Details.ServiceLevels = append(result.Details.ServiceLevels, model.FeedVersionServiceLevel{FeedVersionServiceLevel: v})
	}
	for _, v := range r.Details.Agencies {
		result.Details.Agencies = append(result.Details.Agencies, model.Agency{Agency: v})
	}
	for _, v := range r.Details.Routes {
		result.Details.Routes = append(result.Details.Routes, model.Route{Route: v})
	}
	for _, v := range r.Details.Stops {
		result.Details.Stops = append(result.Details.Stops, model.Stop{Stop: v})
	}
	for _, v := range r.Details.Realtime {
		result.Details.Realtime = append(result.Details.Realtime, model.ValidationRealtimeResult{
			Url:  v.Url,
			Json: v.Json,
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
