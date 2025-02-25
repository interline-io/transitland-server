package gql

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/interline-io/transitland-lib/gtfs"
	"github.com/interline-io/transitland-lib/tldb/postgres"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
)

// Entity mutation tests

func TestStopCreate(t *testing.T) {
	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		finder := cfg.Finder
		ctx := model.WithConfig(context.Background(), cfg)
		fv := model.FeedVersionInput{ID: toPtr(1)}
		stopInput := model.StopSetInput{
			FeedVersion: &fv,
			StopID:      toPtr(fmt.Sprintf("%d", time.Now().UnixNano())),
			StopName:    toPtr("hello"),
			Geometry:    toPtr(tt.NewPoint(-122.271604, 37.803664)),
		}
		eid, err := finder.StopCreate(ctx, stopInput)
		if err != nil {
			t.Fatal(err)
		}
		checkEnt := gtfs.Stop{}
		checkEnt.ID = eid
		atx := postgres.NewPostgresAdapterFromDBX(cfg.Finder.DBX())
		if err := atx.Find(ctx, &checkEnt); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, stopInput.StopID, &checkEnt.StopID.Val)
		assert.Equal(t, stopInput.StopName, &checkEnt.StopName.Val)
		assert.Equal(t, stopInput.Geometry.FlatCoords(), checkEnt.Geometry.FlatCoords())
	})
}

func TestStopUpdate(t *testing.T) {
	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		finder := cfg.Finder
		ctx := model.WithConfig(context.Background(), cfg)
		fv := model.FeedVersionInput{ID: toPtr(1)}
		stopInput := model.StopSetInput{
			FeedVersion: &fv,
			StopID:      toPtr(fmt.Sprintf("%d", time.Now().UnixNano())),
			StopName:    toPtr("hello"),
			Geometry:    toPtr(tt.NewPoint(-122.271604, 37.803664)),
		}
		eid, err := finder.StopCreate(ctx, stopInput)
		if err != nil {
			t.Fatal(err)
		}
		stopUpdate := model.StopSetInput{
			ID:       toPtr(eid),
			StopID:   toPtr(fmt.Sprintf("update-%d", time.Now().UnixNano())),
			Geometry: toPtr(tt.NewPoint(-122.0, 38.0)),
		}
		if _, err := finder.StopUpdate(ctx, stopUpdate); err != nil {
			t.Fatal(err)
		}
		checkEnt := gtfs.Stop{}
		checkEnt.ID = eid
		atx := postgres.NewPostgresAdapterFromDBX(cfg.Finder.DBX())
		if err := atx.Find(ctx, &checkEnt); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, stopUpdate.StopID, &checkEnt.StopID.Val)
		assert.Equal(t, stopUpdate.Geometry.FlatCoords(), checkEnt.Geometry.FlatCoords())
	})
}

func toPtr[T any, P *T](v T) P {
	vcopy := v
	return &vcopy
}
