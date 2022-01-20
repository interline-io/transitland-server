package model

import (
	"time"

	"github.com/interline-io/transitland-lib/rt/pb"

	"github.com/jmoiron/sqlx"
)

type Finder interface {
	FindAgencies(limit *int, after *int, ids []int, where *AgencyFilter) ([]*Agency, error)
	FindRoutes(limit *int, after *int, ids []int, where *RouteFilter) ([]*Route, error)
	FindStops(limit *int, after *int, ids []int, where *StopFilter) ([]*Stop, error)
	FindTrips(limit *int, after *int, ids []int, where *TripFilter) ([]*Trip, error)
	FindFeedVersions(limit *int, after *int, ids []int, where *FeedVersionFilter) ([]*FeedVersion, error)
	FindFeeds(limit *int, after *int, ids []int, where *FeedFilter) ([]*Feed, error)
	FindOperators(limit *int, after *int, ids []int, where *OperatorFilter) ([]*Operator, error)
	RouteStopBuffer(*RouteStopBufferParam) ([]*RouteStopBuffer, error)
	DBX() sqlx.Ext // escape hatch, for now
}

// RTFinder manages and looks up RT data
type RTFinder interface {
	AddData(string, []byte) error
	GetTrip(string, string) (*pb.TripUpdate, bool)
	GetAddedTripsForStop(string, string) []*pb.TripUpdate
	TripGTFSTripID(int) (string, bool)
	FeedVersionOnestopID(int) (string, bool)
	StopTimezone(id int, known string) (*time.Location, bool)
}
