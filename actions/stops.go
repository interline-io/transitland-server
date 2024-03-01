package actions

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/model"
)

func CreateStop(ctx context.Context, input model.StopInput) (int, error) {
	if input.FeedVersionID == nil {
		return 0, errors.New("feed_version_id required")
	}
	fvid := *input.FeedVersionID
	cfg := model.ForContext(ctx)
	dbf := cfg.Finder
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return 0, err
	}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	entId := 0
	err := db.Tx(func(atx tldb.Adapter) error {
		ent := tl.Stop{}
		ent.FeedVersionID = fvid
		if input.StopID != nil {
			ent.StopID = *input.StopID
		}
		if input.StopName != nil {
			ent.StopName = *input.StopName
		}
		if input.LocationType != nil {
			ent.LocationType = *input.LocationType
		}
		if input.Geometry != nil {
			ent.Geometry = *input.Geometry
		}
		var err error
		entId, err = atx.Insert(&ent)
		return err
	})
	if err != nil {
		return 0, err
	}
	return entId, nil
}

func UpdateStop(ctx context.Context, input model.StopInput) (int, error) {
	if input.FeedVersionID == nil {
		return 0, errors.New("feed_version_id required")
	}
	fvid := *input.FeedVersionID
	cfg := model.ForContext(ctx)
	dbf := cfg.Finder
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return 0, err
	}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	entId := 0
	err := db.Tx(func(atx tldb.Adapter) error {
		ent := tl.Stop{}
		var cols []string
		if input.StopID != nil {
			ent.StopID = *input.StopID
			cols = append(cols, "stop_id")
		}
		if input.StopCode != nil {
			ent.StopCode = *input.StopCode
			cols = append(cols, "stop_code")
		}
		if input.StopDesc != nil {
			ent.StopDesc = *input.StopDesc
			cols = append(cols, "stop_desc")
		}
		if input.StopTimezone != nil {
			ent.StopTimezone = *input.StopTimezone
			cols = append(cols, "stop_timezone")
		}
		if input.StopName != nil {
			ent.StopName = *input.StopName
			cols = append(cols, "stop_name")
		}
		if input.StopURL != nil {
			ent.StopURL = *input.StopURL
			cols = append(cols, "stop_url")
		}
		if input.LocationType != nil {
			ent.LocationType = *input.LocationType
			cols = append(cols, "location_type")
		}
		if input.Geometry != nil {
			ent.Geometry = *input.Geometry
			cols = append(cols, "geometry")
		}
		return atx.Update(&ent, cols...)
	})
	if err != nil {
		return 0, err
	}
	return entId, nil
}
