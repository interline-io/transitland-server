package dbfinder

import (
	"context"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/gtfs"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-mw/auth/authz"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) StopCreate(ctx context.Context, input model.StopSetInput) (int, error) {
	if input.FeedVersion == nil || input.FeedVersion.ID == nil {
		return 0, errors.New("feed_version_id required")
	}
	return createUpdateStop(ctx, input)
}

func (f *Finder) StopUpdate(ctx context.Context, input model.StopSetInput) (int, error) {
	if input.ID == nil {
		return 0, errors.New("id required")
	}
	return createUpdateStop(ctx, input)
}

func (f *Finder) StopDelete(ctx context.Context, id int) error {
	ent := gtfs.Stop{}
	ent.ID = id
	return deleteEnt(ctx, &ent)
}

func createUpdateStop(ctx context.Context, input model.StopSetInput) (int, error) {
	return createUpdateEnt(
		ctx,
		input.ID,
		fvint(input.FeedVersion),
		&gtfs.Stop{},
		func(ent *gtfs.Stop) ([]string, error) {
			var cols []string
			cols = checkCol(&ent.StopID, input.StopID, "stop_id", cols)
			cols = checkCol(&ent.StopCode, input.StopCode, "stop_code", cols)
			cols = checkCol(&ent.StopDesc, input.StopDesc, "stop_desc", cols)
			cols = checkCol(&ent.StopTimezone, input.StopTimezone, "stop_timezone", cols)
			cols = checkCol(&ent.StopName, input.StopName, "stop_name", cols)
			cols = checkCol(&ent.StopURL, input.StopURL, "stop_url", cols)
			cols = checkCol(&ent.LocationType, input.LocationType, "location_type", cols)
			cols = checkCol(&ent.WheelchairBoarding, input.WheelchairBoarding, "wheelchair_boarding", cols)
			cols = checkCol(&ent.ZoneID, input.ZoneID, "zone_id", cols)
			cols = scanCol(&ent.TtsStopName, input.TtsStopName, "tts_stop_name", cols)
			cols = scanCol(&ent.PlatformCode, input.PlatformCode, "platform_code", cols)
			if input.Geometry != nil && input.Geometry.Valid {
				cols = checkCol(&ent.Geometry, input.Geometry, "geometry", cols)
			}
			if v := input.Parent; v != nil {
				cols = append(cols, "parent_station")
				scanCol(&ent.ParentStation, v.ID, "parent_station", cols)
			}
			if v := input.Level; v != nil {
				cols = append(cols, "level_id")
				scanCol(&ent.LevelID, v.ID, "level_id", cols)
			}
			return cols, nil
		})
}

///////////

func (f *Finder) PathwayCreate(ctx context.Context, input model.PathwaySetInput) (int, error) {
	if input.FeedVersion == nil || input.FeedVersion.ID == nil {
		return 0, errors.New("feed_version_id required")
	}
	return createUpdatePathway(ctx, input)
}

func (f *Finder) PathwayUpdate(ctx context.Context, input model.PathwaySetInput) (int, error) {
	if input.ID == nil {
		return 0, errors.New("id required")
	}
	return createUpdatePathway(ctx, input)
}

func (f *Finder) PathwayDelete(ctx context.Context, id int) error {
	ent := gtfs.Pathway{}
	ent.ID = id
	return deleteEnt(ctx, &ent)
}

func createUpdatePathway(ctx context.Context, input model.PathwaySetInput) (int, error) {
	return createUpdateEnt(
		ctx,
		input.ID,
		fvint(input.FeedVersion),
		&gtfs.Pathway{},
		func(ent *gtfs.Pathway) ([]string, error) {
			var cols []string
			cols = scanCol(&ent.PathwayID, input.PathwayID, "pathway_id", cols)
			cols = scanCol(&ent.PathwayMode, input.PathwayMode, "pathway_mode", cols)
			cols = scanCol(&ent.IsBidirectional, input.IsBidirectional, "is_bidirectional", cols)
			cols = scanCol(&ent.Length, input.Length, "length", cols)
			cols = scanCol(&ent.TraversalTime, input.TraversalTime, "traversal_time", cols)
			cols = scanCol(&ent.StairCount, input.StairCount, "stair_count", cols)
			cols = scanCol(&ent.MaxSlope, input.MaxSlope, "max_slope", cols)
			cols = scanCol(&ent.MinWidth, input.MinWidth, "min_width", cols)
			cols = scanCol(&ent.SignpostedAs, input.SignpostedAs, "signposted_as", cols)
			cols = scanCol(&ent.ReverseSignpostedAs, input.ReverseSignpostedAs, "reverse_signposted_as", cols)
			if v := input.FromStop; v != nil {
				cols = append(cols, "from_stop_id")
				scanCol(&ent.FromStopID, v.ID, "", nil)
			}
			if v := input.ToStop; v != nil {
				cols = append(cols, "from_stop_id")
				scanCol(&ent.ToStopID, v.ID, "", nil)
			}
			return cols, nil
		})
}

///////////

func (f *Finder) LevelCreate(ctx context.Context, input model.LevelSetInput) (int, error) {
	if input.FeedVersion == nil || input.FeedVersion.ID == nil {
		return 0, errors.New("feed_version_id required")
	}
	return createUpdateLevel(ctx, input)
}

func (f *Finder) LevelUpdate(ctx context.Context, input model.LevelSetInput) (int, error) {
	if input.ID == nil {
		return 0, errors.New("id required")
	}
	return createUpdateLevel(ctx, input)
}

func (f *Finder) LevelDelete(ctx context.Context, id int) error {
	ent := gtfs.Level{}
	ent.ID = id
	return deleteEnt(ctx, &ent)
}

func createUpdateLevel(ctx context.Context, input model.LevelSetInput) (int, error) {
	return createUpdateEnt(
		ctx,
		input.ID,
		fvint(input.FeedVersion),
		&model.Level{},
		func(ent *model.Level) ([]string, error) {
			var cols []string
			cols = scanCol(&ent.LevelID, input.LevelID, "level_id", cols)
			cols = scanCol(&ent.LevelName, input.LevelName, "level_name", cols)
			cols = scanCol(&ent.LevelIndex, input.LevelIndex, "level_index", cols)
			cols = checkCol(&ent.Geometry, input.Geometry, "geometry", cols)
			if v := input.Parent; v != nil {
				cols = append(cols, "parent_station")
				scanCol(&ent.ParentStation, v.ID, "", nil)
			}
			return cols, nil
		})
}

///////////

func checkCol[T any, P *T](val P, inval P, colname string, cols []string) []string {
	if inval != nil {
		*val = *inval
		cols = append(cols, colname)
	}
	return cols
}

type canScan interface {
	Scan(any) error
}

func scanCol[T any, PT *T](val canScan, inval PT, colname string, cols []string) []string {
	if inval != nil {
		if err := val.Scan(*inval); err != nil {
			panic(err)
		}
		cols = append(cols, colname)
	}
	return cols
}

type hasTableName interface {
	TableName() string
	GetFeedVersionID() int
	SetFeedVersionID(int)
	GetID() int
	SetID(int)
	Errors() []error
}

func fvint(fvi *model.FeedVersionInput) *int {
	if fvi == nil {
		return nil
	}
	return fvi.ID
}

func toAtx(ctx context.Context) tldb.Adapter {
	return tldb.NewPostgresAdapterFromDBX(model.ForContext(ctx).Finder.DBX())
}

// ensure we have edit rights to fvid
func createUpdateEnt[T hasTableName](
	ctx context.Context,
	entId *int,
	fvid *int,
	baseEnt T,
	updateFunc func(baseEnt T) ([]string, error),
) (int, error) {
	update := (entId != nil && *entId > 0)
	atx := toAtx(ctx)
	retId := 0

	// Update or create?
	if update {
		baseEnt.SetID(*entId)
		if err := atx.Find(baseEnt); err != nil {
			return 0, err
		}
	} else if fvid != nil {
		baseEnt.SetFeedVersionID(*fvid)
	} else {
		return 0, errors.New("id or feed_version_id required")
	}

	// Check we can edit this feed version
	if err := checkFeedEdit(ctx, baseEnt.GetFeedVersionID()); err != nil {
		return 0, err
	}

	// Update columns
	cols, err := updateFunc(baseEnt)
	if err != nil {
		return 0, err
	}

	// Validate
	if errs := baseEnt.Errors(); len(errs) > 0 {
		return 0, errs[0]
	}
	// Save
	err = nil
	if update {
		retId = baseEnt.GetID()
		err = atx.Update(baseEnt, cols...)
	} else {
		retId, err = atx.Insert(baseEnt)
	}
	if err != nil {
		return 0, err
	}
	return retId, nil

}

// ensure we have edit rights to fvid
func deleteEnt(ctx context.Context, ent hasTableName) error {
	// Get fvid
	entId := ent.GetID()
	fvid := ent.GetFeedVersionID()
	db := model.ForContext(ctx).Finder.DBX()
	if err := dbutil.Get(
		ctx,
		db,
		sq.StatementBuilder.
			Select("feed_version_id").
			From(ent.TableName()).
			Where(sq.Eq{"id": entId}),
		&fvid,
	); err != nil {
		return err
	}

	// // Check if we can edit fv
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return err
	}

	// Perform deletion
	// TODO: use dbutil.Delete
	atx := toAtx(ctx)
	_, err := atx.Sqrl().Delete(ent.TableName()).Where(sq.Eq{"id": entId}).Exec()
	return err
}

func checkFeedEdit(ctx context.Context, fvid int) error {
	if fvid <= 0 {
		return errors.New("invalid feed version id")
	}
	cfg := model.ForContext(ctx)
	if checker := cfg.Checker; checker == nil {
		return nil
	} else if check, err := checker.FeedVersionPermissions(ctx, &authz.FeedVersionRequest{Id: int64(fvid)}); err != nil {
		return err
	} else if !check.Actions.CanEdit {
		return authz.ErrUnauthorized
	}
	return nil
}
