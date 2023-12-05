package actions

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/causes"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tlcsv"
	"github.com/interline-io/transitland-lib/validator"
	"github.com/interline-io/transitland-server/auth/authn"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
)

type hasGeometries interface {
	Geometries() []tt.Geometry
}

type ValidationReport struct {
	tl.BaseEntity
	ReportedAt tt.Time
}

func (e *ValidationReport) TableName() string {
	return "tl_validation_reports"
}

type ValidationReportTripUpdateStat struct {
	ValidationReportID int
	RouteID            string
	TripScheduledCount int
	TripMatchCount     int
}

// ValidateUpload takes a file Reader and produces a validation package containing errors, warnings, file infos, service levels, etc.
func ValidateUpload(ctx context.Context, cfg config.Config, dbf model.Finder, src io.Reader, feedURL *string, rturls []string, user authn.User) (*model.ValidationResult, error) {
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
		IncludeRealtimeJson:      true,
		IncludeEntitiesLimit:     10_000,
		EvaluateAt:               time.Date(2018, 1, 18, 16, 0, 0, 0, time.UTC),
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

	// validationReport := ValidationReport{}
	// validationReport.FeedVersionID = 1
	// validationReport.ReportedAt = tt.NewTime(time.Now())
	// db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	// if eid, err := db.Insert(&validationReport); err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("eid:", eid)
	// }

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
			Field:     eg.Field,
			ErrorCode: eg.ErrorCode,
			ErrorType: eg.ErrorType,
			Message:   eg.Message,
			Count:     eg.Count,
			Limit:     eg.Limit,
		}
		for _, err := range eg.Errors {
			err2 := model.ValidationResultError{
				Filename:  eg.Filename,
				Message:   err.Error(),
				ErrorType: eg.ErrorType,
				ErrorCode: eg.ErrorCode,
			}
			if v, ok := err.(hasContext); ok {
				c := v.Context()
				err2.EntityID = c.EntityID
				err2.ErrorCode = c.Code
				err2.Field = c.Field
			}
			if v, ok := err.(hasGeometries); ok {
				for _, g := range v.Geometries() {
					g := g
					err2.Geometries = append(err2.Geometries, &g)
				}
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
			Field:     eg.Field,
			ErrorCode: eg.ErrorCode,
			ErrorType: eg.ErrorType,
			Message:   eg.Message,
			Count:     eg.Count,
			Limit:     eg.Limit,
		}

		for _, err := range eg.Errors {
			err2 := model.ValidationResultError{
				Filename:  eg.Filename,
				Message:   err.Error(),
				ErrorType: eg.ErrorType,
				ErrorCode: eg.ErrorCode,
			}
			if v, ok := err.(hasContext); ok {
				c := v.Context()
				err2.EntityID = c.EntityID
				err2.Field = c.Field
			}
			if v, ok := err.(hasGeometries); ok {
				for _, g := range v.Geometries() {
					g := g
					err2.Geometries = append(err2.Geometries, &g)
				}
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
		result.Routes = append(result.Routes, model.Route{Route: v})
	}
	for _, v := range r.Stops {
		result.Stops = append(result.Stops, model.Stop{Stop: v})
	}
	for _, v := range r.Realtime {
		result.Realtime = append(result.Realtime, model.ValidationRealtimeResult{
			Url:  v.Url,
			Json: v.Json,
		})
		for _, rs := range v.VehiclePositionStats {
			_ = rs
		}
		for _, rs := range v.TripUpdateStats {
			fmt.Printf("RS: %#v\n", rs)
			_ = rs
		}
	}
	return &result, nil
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
