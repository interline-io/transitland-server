package testfinder

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/interline-io/transitland-lib/rt/pb"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Test helpers

var db *sqlx.DB

type TestEnv struct {
	Config     config.Config
	Finder     model.Finder
	RTFinder   model.RTFinder
	GbfsFinder model.GbfsFinder
}

func Finders(t testing.TB, cl clock.Clock, rtJsons []RTJsonFile) TestEnv {
	g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
	if g == "" {
		t.Fatal("TL_TEST_SERVER_DATABASE_URL not set, skipping")
	}
	if cl == nil {
		cl = &clock.Real{}
	}
	cfg := config.Config{
		Clock:     cl,
		Storage:   t.TempDir(),
		RTStorage: t.TempDir(),
	}
	if db == nil {
		db = dbfinder.MustOpenDB(g)
	}
	dbf := dbfinder.NewDBFinder(db)
	dbf.Clock = cl
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	rtf.Clock = cl
	gbf := gbfsfinder.NewFinder(nil)
	for _, rtj := range rtJsons {
		fn := testutil.RelPath("test", "data", "rt", rtj.Fname)
		if err := FetchRTJson(rtj.Feed, rtj.Ftype, fn, rtf); err != nil {
			t.Fatal(err)
		}
	}
	return TestEnv{
		Config:     cfg,
		Finder:     dbf,
		RTFinder:   rtf,
		GbfsFinder: gbf,
	}
}

func FindersTx(t testing.TB, cl clock.Clock, rtJsons []RTJsonFile, cb func(TestEnv) error) {
	// Check open DB
	if db == nil {
		g := os.Getenv("TL_TEST_SERVER_DATABASE_URL")
		if g == "" {
			t.Fatal("TL_TEST_SERVER_DATABASE_URL not set, skipping")
		}
		db = dbfinder.MustOpenDB(g)
	}

	// Config
	if cl == nil {
		cl = &clock.Real{}
	}
	cfg := config.Config{
		Clock:     cl,
		Storage:   t.TempDir(),
		RTStorage: t.TempDir(),
	}

	// Start Txn
	tx := db.MustBeginTx(context.Background(), nil)
	defer tx.Rollback()

	// Configure finders
	dbf := dbfinder.NewDBFinder(tx)
	dbf.Clock = cl
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), tx)
	rtf.Clock = cl
	gbf := gbfsfinder.NewFinder(nil)
	for _, rtj := range rtJsons {
		fn := testutil.RelPath("test", "data", "rt", rtj.Fname)
		if err := FetchRTJson(rtj.Feed, rtj.Ftype, fn, rtf); err != nil {
			t.Fatal(err)
		}
	}

	testEnv := TestEnv{
		Config:     cfg,
		Finder:     dbf,
		RTFinder:   rtf,
		GbfsFinder: gbf,
	}

	// Commit or rollback
	if err := cb(testEnv); err != nil {
		//tx.Rollback()
	} else {
		tx.Commit()
	}
}

func FindersTxRollback(t testing.TB, cl clock.Clock, rtJsons []RTJsonFile, cb func(TestEnv)) {
	FindersTx(t, cl, rtJsons, func(c TestEnv) error {
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
