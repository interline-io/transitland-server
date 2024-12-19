package testconfig

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/interline-io/transitland-dbutil/testutil"
	"github.com/interline-io/transitland-jobs/jobs"
	localjobs "github.com/interline-io/transitland-jobs/local"
	"github.com/interline-io/transitland-lib/rt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/finders/actions"
	"github.com/interline-io/transitland-server/finders/azchecker"
	"github.com/interline-io/transitland-server/finders/dbfinder"
	"github.com/interline-io/transitland-server/finders/gbfsfinder"
	"github.com/interline-io/transitland-server/finders/rtfinder"
	"github.com/interline-io/transitland-server/internal/clock"
	"github.com/interline-io/transitland-server/model"
	"github.com/interline-io/transitland-server/testdata"
	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"
)

// Test helpers

type Options struct {
	WhenUtc        string
	Storage        string
	RTStorage      string
	RTJsons        []RTJsonFile
	FGAEndpoint    string
	FGAModelFile   string
	FGAModelTuples []authz.TupleKey
}

func Config(t testing.TB, opts Options) model.Config {
	db := testutil.MustOpenTestDB(t)
	return newTestConfig(t, &tldb.QueryLogger{Ext: db}, opts)
}

func ConfigTx(t testing.TB, opts Options, cb func(model.Config) error) {
	// Start Txn
	db := testutil.MustOpenTestDB(t)
	tx := db.MustBeginTx(context.Background(), nil)
	defer tx.Rollback()

	// Get finders
	testEnv := newTestConfig(t, &tldb.QueryLogger{Ext: tx}, opts)

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
	if opts.WhenUtc == "" {
		opts.WhenUtc = "2022-09-01T00:00:00Z"
	}

	when, err := time.Parse("2006-01-02T15:04:05Z", opts.WhenUtc)
	if err != nil {
		t.Fatal(err)
	}
	cl := &clock.Mock{T: when}

	// Setup Checker
	var checker model.Checker
	if opts.FGAEndpoint != "" {
		checkerCfg := azchecker.CheckerConfig{
			FGAEndpoint:      opts.FGAEndpoint,
			FGALoadModelFile: opts.FGAModelFile,
			FGALoadTestData:  opts.FGAModelTuples,
		}
		checker, err = azchecker.NewCheckerFromConfig(checkerCfg, db)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Setup DB
	dbf := dbfinder.NewFinder(db)
	dbf.Clock = cl

	// Setup RT
	ctx := context.Background()
	rtf := rtfinder.NewFinder(rtfinder.NewLocalCache(), db)
	rtf.Clock = cl
	for _, rtj := range opts.RTJsons {
		fn := testdata.Path("rt", rtj.Fname)
		msg, err := rt.ReadFile(fn)
		if err != nil {
			t.Fatal(err)
		}
		key := fmt.Sprintf("rtdata:%s:%s", rtj.Feed, rtj.Ftype)
		rtdata, err := proto.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		if err := rtf.AddData(ctx, key, rtdata); err != nil {
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

	// Initialize job queue - do not start
	jobQueue := jobs.NewJobLogger(localjobs.NewLocalJobs())

	// Action finder
	actionFinder := &actions.Actions{}

	return model.Config{
		Finder:     dbf,
		RTFinder:   rtf,
		GbfsFinder: gbf,
		Checker:    checker,
		JobQueue:   jobQueue,
		Actions:    actionFinder,
		Clock:      cl,
		Storage:    opts.Storage,
		RTStorage:  opts.RTStorage,
	}
}
