package testconfig

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/interline-io/transitland-lib/rt"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-mw/auth/azcheck"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/internal/testutil"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

// Test helpers

type Options struct {
	When           string
	Storage        string
	RTStorage      string
	RTJsons        []RTJsonFile
	FGAModelFile   string
	FGAModelTuples []authz.TupleKey
}

func Config(t testing.TB, opts Options) model.Config {
	db := testutil.MustOpenTestDB()
	return newTestConfig(t, db, opts)
}

func ConfigTx(t testing.TB, opts Options, cb func(model.Config) error) {
	// Check open DB
	db := testutil.MustOpenTestDB()

	// Start Txn
	tx := db.MustBeginTx(context.Background(), nil)
	defer tx.Rollback()

	// Get finders
	testEnv := newTestConfig(t, tx, opts)

	// Commit or rollback
	if err := cb(testEnv); err != nil {
		//tx.Rollback()
	} else {
		tx.Commit()
	}
}

func ConfigTxRollback(t testing.TB, opts Options, cb func(model.Config)) {
	ConfigTx(t, opts, func(c model.Config) error {
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

func newTestConfig(t testing.TB, db sqlx.Ext, opts Options) model.Config {
	// Default time
	if opts.When == "" {
		opts.When = "2022-09-01T00:00:00"
	}

	when, err := time.Parse("2006-01-02T15:04:05", opts.When)
	if err != nil {
		t.Fatal(err)
	}
	cl := &clock.Mock{T: when}

	// Setup Checker
	checkerCfg := azcheck.CheckerConfig{
		FGAEndpoint:      os.Getenv("TL_TEST_FGA_ENDPOINT"),
		FGALoadModelFile: opts.FGAModelFile,
		FGALoadTestData:  opts.FGAModelTuples,
	}
	checker, err := azcheck.NewCheckerFromConfig(checkerCfg, db)
	if err != nil {
		t.Fatal(err)
	}

	// Setup DB
	dbf := dbfinder.NewFinder(db)
	dbf.Clock = cl

	// Setup RT
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	rtf.Clock = cl
	for _, rtj := range opts.RTJsons {
		fn := testutil.RelPath("test", "data", "rt", rtj.Fname)
		msg, err := rt.ReadFile(fn)
		if err != nil {
			t.Fatal(err)
		}
		key := fmt.Sprintf("rtdata:%s:%s", rtj.Feed, rtj.Ftype)
		rtdata, err := proto.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		if err := rtf.AddData(key, rtdata); err != nil {
			t.Fatal(err)
		}
	}

	// Setup GBFS
	gbf := gbfsfinder.NewFinder(nil)

	if opts.Storage == "" {
		opts.Storage = t.TempDir()
	}
	if opts.RTStorage == "" {
		opts.RTStorage = t.TempDir()
	}
	return model.Config{
		Finder:     dbf,
		RTFinder:   rtf,
		GbfsFinder: gbf,
		Checker:    checker,
		Clock:      cl,
		Storage:    opts.Storage,
		RTStorage:  opts.RTStorage,
	}
}
