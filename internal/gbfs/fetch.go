package gbfs

import (
	"encoding/json"

	"github.com/interline-io/transitland-lib/dmfr/fetch"
	"github.com/interline-io/transitland-lib/tl/request"
)

type Options struct {
	Language string
	fetch.Options
}

type Result struct {
	fetch.Result
}

func Fetch(opts Options) ([]GbfsFeed, Result, error) {
	res := Result{}
	var reqOpts []request.RequestOption
	if opts.AllowFTPFetch {
		reqOpts = append(reqOpts, request.WithAllowFTP)
	}
	if opts.AllowLocalFetch {
		reqOpts = append(reqOpts, request.WithAllowLocal)
	}
	if opts.AllowS3Fetch {
		reqOpts = append(reqOpts, request.WithAllowS3)
	}
	systemFile := SystemFile{}
	fr, err := fetchUnmarshal(opts.FeedURL, &systemFile, reqOpts...)
	res.ResponseCode = fr.ResponseCode
	res.ResponseSHA1 = fr.ResponseSHA1
	res.ResponseSize = fr.ResponseSize
	if err != nil {
		return nil, res, err
	}
	var feeds []GbfsFeed
	for _, sflang := range systemFile.Data {
		feed, err := fetchAll(sflang)
		if err == nil {
			feeds = append(feeds, feed)
		}
	}
	return feeds, res, nil
}

func fetchAll(sf SystemFeeds, reqOpts ...request.RequestOption) (GbfsFeed, error) {
	ret := GbfsFeed{}
	var err error
	for _, v := range sf.Feeds {
		switch v.Name.Val {
		case "system_information":
			e := SystemInformationFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.SystemInformation = e.Data
		case "station_information":
			e := StationInformationFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.StationInformation = e.Data.Stations
		case "station_status":
			e := StationStatusFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.StationStatus = e.Data.Stations
		case "free_bike_status":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.Bikes = e.Bikes
		case "system_hours":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.RentalHours = e.RentalHours
		case "system_calendar":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.Calendars = e.Calendars
		case "system_regions":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.Regions = e.Regions
		case "system_alerts":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.Alerts = e.Alerts
		case "vehicle_types":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.VehicleTypes = e.VehicleTypes
		case "system_pricing_plans":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.Plans = e.Plans
		case "geofencing_zones":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.GeofencingZones = e.GeofencingZones
		case "gbfs_versions":
			e := GbfsFile{}
			_, err = fetchUnmarshal(v.URL.Val, &e, reqOpts...)
			ret.Versions = e.Versions
		}
	}
	return ret, err
}

func fetchUnmarshal(url string, ent any, reqOpts ...request.RequestOption) (request.FetchResponse, error) {
	fr, err := request.AuthenticatedRequest(url, reqOpts...)
	if err != nil {
		return fr, err
	}
	if err := json.Unmarshal(fr.Data, ent); err != nil {
		return fr, err
	}
	return fr, nil
}
