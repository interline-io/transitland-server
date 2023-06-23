package testfinder

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/auth/authz"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Test helpers

var db *sqlx.DB

type TestFinderOptions struct {
	Clock          clock.Clock
	RTJsons        []RTJsonFile
	FGAModelFile   string
	FGAModelTuples []authz.TupleKey
}

func newFinders(t testing.TB, db sqlx.Ext, opts TestFinderOptions) model.Finders {
	if opts.Clock == nil {
		opts.Clock = &clock.Real{}
	}
	cfg := config.Config{
		Clock:     opts.Clock,
		Storage:   t.TempDir(),
		RTStorage: t.TempDir(),
	}

	// Setup DB
	dbf := dbfinder.NewFinder(db)
	dbf.Clock = opts.Clock

	// Setup RT
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	rtf.Clock = opts.Clock

	// Setup GBFS
	gbf := gbfsfinder.NewFinder(nil)
	for _, rtj := range opts.RTJsons {
		fn := testutil.RelPath("test", "data", "rt", rtj.Fname)
		if err := FetchRTJson(rtj.Feed, rtj.Ftype, fn, rtf); err != nil {
			t.Fatal(err)
		}
	}

	// Setup Checker
	checkerCfg := authz.AuthzConfig{
		FGAEndpoint:      os.Getenv("TL_TEST_FGA_ENDPOINT"),
		FGALoadModelFile: opts.FGAModelFile,
		FGALoadTestData:  opts.FGAModelTuples,
	}
	checker, err := authz.NewCheckerFromConfig(checkerCfg, db, nil)
	if err != nil {
		t.Fatal(err)
	}
	return model.Finders{
		Config:     cfg,
		Finder:     dbf,
		RTFinder:   rtf,
		GbfsFinder: gbf,
		Checker:    checker,
	}
}

func Finders(t testing.TB, cl clock.Clock, rtJsons []RTJsonFile) model.Finders {
	if db == nil {
		db = dbutil.MustOpenTestDB()
	}
	return newFinders(t, db, TestFinderOptions{Clock: cl, RTJsons: rtJsons})
}

func FindersWithOptions(t testing.TB, opts TestFinderOptions) model.Finders {
	if db == nil {
		db = dbutil.MustOpenTestDB()
	}
	return newFinders(t, db, opts)
}

func FindersTx(t testing.TB, cl clock.Clock, rtJsons []RTJsonFile, cb func(model.Finders) error) {
	// Check open DB
	if db == nil {
		db = dbutil.MustOpenTestDB()
	}
	// Start Txn
	tx := db.MustBeginTx(context.Background(), nil)
	defer tx.Rollback()

	// Get finders
	testEnv := newFinders(t, db, TestFinderOptions{Clock: cl, RTJsons: rtJsons})

	// Commit or rollback
	if err := cb(testEnv); err != nil {
		//tx.Rollback()
	} else {
		tx.Commit()
	}
}

func FindersTxRollback(t testing.TB, cl clock.Clock, rtJsons []RTJsonFile, cb func(model.Finders)) {
	FindersTx(t, cl, rtJsons, func(c model.Finders) error {
		cb(c)
		return errors.New("rollback")
	})
}

type RTJsonFile struct {
	Feed  string
	Ftype string
	Fname string
}

func DefaultRTJson() []RTJsonFile {
	return []RTJsonFile{
		{"BA", "realtime_trip_updates", "BA.json"},
		{"BA", "realtime_alerts", "BA-alerts.json"},
		{"CT", "realtime_trip_updates", "CT.json"},
	}
}

// FetchRTJson fetches test protobuf in JSON format
// URL is relative to project root
func FetchRTJson(feed string, ftype string, url string, rtfinder model.RTFinder) error {
	var msg pb.FeedMessage
	jdata, err := ioutil.ReadFile(url)
	if err != nil {
		return err
	}
	if err := protojson.Unmarshal(jdata, &msg); err != nil {
		return err
	}
	rtdata, err := proto.Marshal(&msg)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("rtdata:%s:%s", feed, ftype)
	return rtfinder.AddData(key, rtdata)
}
