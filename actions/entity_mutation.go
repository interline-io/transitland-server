package actions

import (
	"context"
	"errors"
	"strconv"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/model"
	"github.com/jmoiron/sqlx"
)

func CreateStop(ctx context.Context, input model.StopInput) (int, error) {
	if input.FeedVersionID == nil {
		return 0, errors.New("feed_version_id required")
	}
	return createUpdateStop(ctx, input)
}

func UpdateStop(ctx context.Context, input model.StopInput) (int, error) {
	if input.ID == nil {
		return 0, errors.New("id required")
	}
	return createUpdateStop(ctx, input)
}

func createUpdateStop(ctx context.Context, input model.StopInput) (int, error) {
	entId := 0
	update := (input.ID != nil)
	db := model.ForContext(ctx).Finder.DBX()
	err := tldb.NewPostgresAdapterFromDBX(db).Tx(func(atx tldb.Adapter) error {
		ent := tl.Stop{}
		if update {
			if err := getEnt(ctx, db, ent.TableName(), *input.ID, &ent); err != nil {
				return err
			}
		} else {
			ent.FeedVersionID = *input.FeedVersionID
		}
		if err := checkFeedEdit(ctx, ent.FeedVersionID); err != nil {
			return err
		}
		var cols []string
		cols = checkCol(&ent.StopID, input.StopID, "stop_id", cols)
		cols = checkCol(&ent.Geometry, input.Geometry, "geometry", cols)
		cols = checkCol(&ent.StopCode, input.StopCode, "stop_code", cols)
		cols = checkCol(&ent.StopDesc, input.StopDesc, "stop_desc", cols)
		cols = checkCol(&ent.StopTimezone, input.StopTimezone, "stop_timezone", cols)
		cols = checkCol(&ent.StopName, input.StopName, "stop_name", cols)
		cols = checkCol(&ent.StopURL, input.StopURL, "stop_url", cols)
		cols = checkCol(&ent.LocationType, input.LocationType, "location_type", cols)
		cols = checkCol(&ent.WheelchairBoarding, input.WheelchairBoarding, "wheelchair_boarding", cols)
		cols = checkCol(&ent.ZoneID, input.ZoneID, "zone_id", cols)
		// cols = checkCol(&ent.TtsStopName, input.TtsStopName, "tts_stop_name", cols)
		// cols = checkCol(&ent.PlatformCode, input.PlatformCode, "stop_url", cols)
		if v := input.Parent; v != nil {
			if v.ID == nil {
				ent.ParentStation.Valid = false
			} else if err := matchFvid(ctx, db, ent.FeedVersionID, "gtfs_stops", *v.ID); err != nil {
				return err
			} else {
				ent.ParentStation = tt.NewKey(strconv.Itoa(*v.ID))
				cols = append(cols, "parent_station")
			}
		}
		if v := input.Level; v != nil {
			if v.ID == nil {
				ent.LevelID.Valid = false
			} else if err := matchFvid(ctx, db, ent.FeedVersionID, "gtfs_levels", *v.ID); err != nil {
				return err
			} else {
				ent.LevelID = tt.NewKey(strconv.Itoa(*v.ID))
				cols = append(cols, "level_id")
			}
		}

		// Validate
		if errs := ent.Errors(); len(errs) > 0 {
			return errs[0]
		}
		// Save
		if update {
			entId = ent.ID
			return atx.Update(&ent, cols...)
		} else {
			var err error
			entId, err = atx.Insert(&ent)
			return err
		}
	})
	if err != nil {
		return 0, err
	}
	return entId, nil
}

///////////

func CreatePathway(ctx context.Context, input model.PathwayInput) (int, error) {
	if input.FeedVersionID == nil {
		return 0, errors.New("feed_version_id required")
	}
	return createUpdatePathway(ctx, input)
}

func UpdatePathway(ctx context.Context, input model.PathwayInput) (int, error) {
	if input.ID == nil {
		return 0, errors.New("id required")
	}
	return createUpdatePathway(ctx, input)
}

func createUpdatePathway(ctx context.Context, input model.PathwayInput) (int, error) {
	entId := 0
	update := (input.ID != nil)
	db := model.ForContext(ctx).Finder.DBX()
	err := tldb.NewPostgresAdapterFromDBX(db).Tx(func(atx tldb.Adapter) error {
		ent := tl.Pathway{}
		if update {
			if err := getEnt(ctx, db, ent.TableName(), *input.ID, &ent); err != nil {
				return err
			}
		} else {
			ent.FeedVersionID = *input.FeedVersionID
		}
		if err := checkFeedEdit(ctx, ent.FeedVersionID); err != nil {
			return err
		}
		var cols []string
		cols = checkCol(&ent.PathwayID, input.PathwayID, "pathway_id", cols)
		cols = checkCol(&ent.PathwayMode, input.PathwayMode, "pathway_mode", cols)
		cols = checkCol(&ent.IsBidirectional, input.IsBidirectional, "is_bidirectional", cols)
		cols = checkCol(&ent.Length, input.Length, "length", cols)
		cols = checkCol(&ent.TraversalTime, input.TraversalTime, "traversal_time", cols)
		cols = checkCol(&ent.StairCount, input.StairCount, "stair_count", cols)
		cols = checkCol(&ent.MaxSlope, input.MaxSlope, "max_slope", cols)
		cols = checkCol(&ent.MinWidth, input.MinWidth, "min_width", cols)
		cols = checkCol(&ent.SignpostedAs, input.SignpostedAs, "signposted_as", cols)
		cols = checkCol(&ent.ReverseSignpostedAs, input.ReverseSignpostedAs, "reverse_signposted_as", cols)
		if v := input.FromStop; v != nil {
			if v.ID == nil {
				ent.FromStopID = ""
			} else if err := matchFvid(ctx, db, ent.FeedVersionID, "gtfs_stops", *v.ID); err != nil {
				return err
			} else {
				ent.FromStopID = strconv.Itoa(*v.ID)
				cols = append(cols, "from_stop_id")
			}
		}
		if v := input.ToStop; v != nil {
			if v.ID == nil {
				ent.ToStopID = ""
			} else if err := matchFvid(ctx, db, ent.FeedVersionID, "gtfs_stops", *v.ID); err != nil {
				return err
			} else {
				ent.ToStopID = strconv.Itoa(*v.ID)
				cols = append(cols, "to_stop_id")
			}
		}

		// Validate
		if errs := ent.Errors(); len(errs) > 0 {
			return errs[0]
		}
		// Save
		if update {
			entId = ent.ID
			return atx.Update(&ent, cols...)
		} else {
			var err error
			entId, err = atx.Insert(&ent)
			return err
		}
	})
	if err != nil {
		return 0, err
	}
	return entId, nil
}

///////////

func CreateLevel(ctx context.Context, input model.LevelInput) (int, error) {
	if input.FeedVersionID == nil {
		return 0, errors.New("feed_version_id required")
	}
	return createUpdateLevel(ctx, input)
}

func UpdateLevel(ctx context.Context, input model.LevelInput) (int, error) {
	if input.ID == nil {
		return 0, errors.New("id required")
	}
	return createUpdateLevel(ctx, input)
}

func createUpdateLevel(ctx context.Context, input model.LevelInput) (int, error) {
	entId := 0
	update := (input.ID != nil)
	db := model.ForContext(ctx).Finder.DBX()
	err := tldb.NewPostgresAdapterFromDBX(db).Tx(func(atx tldb.Adapter) error {
		ent := model.Level{} // Use model, not tl.Level
		if update {
			if err := getEnt(ctx, db, ent.TableName(), *input.ID, &ent); err != nil {
				return err
			}
		} else {
			ent.FeedVersionID = *input.FeedVersionID
		}
		if err := checkFeedEdit(ctx, ent.FeedVersionID); err != nil {
			return err
		}
		var cols []string
		cols = checkCol(&ent.LevelID, input.LevelID, "level_id", cols)
		cols = checkCol(&ent.LevelName, input.LevelName, "level_name", cols)
		cols = checkCol(&ent.LevelIndex, input.LevelIndex, "level_index", cols)
		cols = checkCol(&ent.Geometry, input.Geometry, "geometry", cols)
		// Validate
		if errs := ent.Errors(); len(errs) > 0 {
			return errs[0]
		}
		// Save
		if update {
			entId = ent.ID
			return atx.Update(&ent, cols...)
		} else {
			var err error
			entId, err = atx.Insert(&ent)
			return err
		}
	})
	if err != nil {
		return 0, err
	}
	return entId, nil
}

///////////

func checkCol[T any, P *T](val P, inval P, colname string, cols []string) []string {
	if inval != nil {
		*val = *inval
		cols = append(cols, colname)
	}
	return cols
}

func getEnt(ctx context.Context, db sqlx.Ext, entTableName string, entId int, ent any) error {
	if err := dbutil.Get(
		ctx,
		db,
		sq.StatementBuilder.Select("*").From(entTableName).Where(sq.Eq{"id": entId}),
		ent,
	); err != nil {
		return err
	}
	return nil
}

func getFvid(ctx context.Context, db sqlx.Ext, entTableName string, entId int) (int, error) {
	getFvid := 0
	if err := dbutil.Get(
		ctx,
		db,
		sq.StatementBuilder.Select("feed_version_id").From(entTableName).Where(sq.Eq{"id": entId}),
		&getFvid,
	); err != nil {
		return getFvid, err
	}
	return getFvid, nil
}

func matchFvid(ctx context.Context, db sqlx.Ext, checkFvid int, entTableName string, entId int) error {
	fvid, err := getFvid(ctx, db, entTableName, entId)
	if err != nil {
		return err
	}
	if checkFvid != fvid {
		return errors.New("mismatched feed_version_id")
	}
	return nil
}

func toPtr[T any, P *T](v T) P {
	vcopy := v
	return &vcopy
}
