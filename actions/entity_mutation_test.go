package actions

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/internal/testconfig"
	"github.com/interline-io/transitland-server/model"
	"github.com/stretchr/testify/assert"
)

func TestCreateStop(t *testing.T) {
	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		ctx := model.WithConfig(context.Background(), cfg)
		stopInput := model.StopInput{
			FeedVersionID: toPtr(1),
			StopID:        toPtr(fmt.Sprintf("%d", time.Now().UnixNano())),
			StopName:      toPtr("hello"),
			Geometry:      toPtr(tt.NewPoint(-122.271604, 37.803664)),
		}
		eid, err := CreateStop(ctx, stopInput)
		if err != nil {
			t.Fatal(err)
		}
		checkEnt := tl.Stop{}
		if err := getEnt(ctx, cfg.Finder.DBX(), "gtfs_stops", eid, &checkEnt); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, stopInput.StopID, &checkEnt.StopID)
		assert.Equal(t, stopInput.StopName, &checkEnt.StopName)
		assert.Equal(t, stopInput.FeedVersionID, &checkEnt.FeedVersionID)
		assert.Equal(t, stopInput.Geometry.Coords(), checkEnt.Geometry.Coords())
	})
}

func TestUpdateStop(t *testing.T) {
	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
		ctx := model.WithConfig(context.Background(), cfg)
		stopInput := model.StopInput{
			FeedVersionID: toPtr(1),
			StopID:        toPtr(fmt.Sprintf("%d", time.Now().UnixNano())),
			StopName:      toPtr("hello"),
			Geometry:      toPtr(tt.NewPoint(-122.271604, 37.803664)),
		}
		eid, err := CreateStop(ctx, stopInput)
		if err != nil {
			t.Fatal(err)
		}
		stopUpdate := model.StopInput{
			ID:       toPtr(eid),
			StopID:   toPtr(fmt.Sprintf("update-%d", time.Now().UnixNano())),
			Geometry: toPtr(tt.NewPoint(-122.0, 38.0)),
		}
		if _, err := UpdateStop(ctx, stopUpdate); err != nil {
			t.Fatal(err)
		}
		checkEnt := tl.Stop{}
		if err := getEnt(ctx, cfg.Finder.DBX(), "gtfs_stops", eid, &checkEnt); err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, stopUpdate.StopID, &checkEnt.StopID)
		assert.Equal(t, stopUpdate.Geometry.Coords(), checkEnt.Geometry.Coords())
	})
}
