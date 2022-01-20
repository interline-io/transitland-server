package find

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// TODO: replace with middleware or configuration

func MustOpenDB(url string) sqlx.Ext {
	db, err := sqlx.Open("postgres", url)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(time.Hour)
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}
	db.Mapper = reflectx.NewMapperFunc("db", toSnakeCase)
	return db.Unsafe()
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// MustSelect runs a query or panics.
func MustSelect(db sqlx.Ext, q sq.SelectBuilder, dest interface{}) {
	q = q.PlaceholderFormat(sq.Dollar)
	qstr, qargs := q.MustSql()
	if os.Getenv("TL_LOG_SQL") == "true" {
		fmt.Println(qstr)
	}
	if a, ok := db.(sqlx.Preparer); ok {
		stmt, err := sqlx.Preparex(a, qstr)
		if err != nil {
			panic(err)
		}
		if err := stmt.Select(dest, qargs...); err != nil {
			panic(err)
		}
	} else {
		if err := sqlx.Select(db, dest, qstr, qargs...); err != nil {
			panic(err)
		}
	}
}

////////

type DBFinder struct {
	db sqlx.Ext
}

func NewDBFinder(db sqlx.Ext) *DBFinder {
	return &DBFinder{db: db}
}

func (f *DBFinder) DBX() sqlx.Ext {
	return f.db
}

func (f *DBFinder) FindAgencies(limit *int, after *int, ids []int, where *model.AgencyFilter) ([]*model.Agency, error) {
	var ents []*model.Agency
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := AgencySelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindRoutes(limit *int, after *int, ids []int, where *model.RouteFilter) ([]*model.Route, error) {
	var ents []*model.Route
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := RouteSelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindStops(limit *int, after *int, ids []int, where *model.StopFilter) ([]*model.Stop, error) {
	var ents []*model.Stop
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := StopSelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindTrips(limit *int, after *int, ids []int, where *model.TripFilter) ([]*model.Trip, error) {
	var ents []*model.Trip
	active := false
	if where != nil && where.FeedVersionSha1 == nil && len(ids) == 0 {
		active = true
	}
	q := TripSelect(limit, after, ids, active, where)
	MustSelect(f.db, q, &ents)
	return ents, nil
}

func (f *DBFinder) FindFeedVersions(limit *int, after *int, ids []int, where *model.FeedVersionFilter) ([]*model.FeedVersion, error) {
	var ents []*model.FeedVersion
	MustSelect(f.db, FeedVersionSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) FindFeeds(limit *int, after *int, ids []int, where *model.FeedFilter) ([]*model.Feed, error) {
	var ents []*model.Feed
	MustSelect(f.db, FeedSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) FindOperators(limit *int, after *int, ids []int, where *model.OperatorFilter) ([]*model.Operator, error) {
	var ents []*model.Operator
	MustSelect(f.db, OperatorSelect(limit, after, ids, where), &ents)
	return ents, nil
}

func (f *DBFinder) RouteStopBuffer(param *model.RouteStopBufferParam) ([]*model.RouteStopBuffer, error) {
	if param == nil {
		return nil, nil
	}
	var ents []*model.RouteStopBuffer
	q := RouteStopBufferSelect(*param)
	MustSelect(f.db, q, &ents)
	return ents, nil
}
