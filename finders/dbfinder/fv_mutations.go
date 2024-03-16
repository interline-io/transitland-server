package dbfinder

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-lib/dmfr/importer"
	"github.com/interline-io/transitland-lib/dmfr/unimporter"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/model"
)

func (f *Finder) FeedVersionImport(ctx context.Context, fvid int) (*model.FeedVersionImportResult, error) {
	cfg := model.ForContext(ctx)
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return nil, err
	}

	opts := importer.Options{
		FeedVersionID: fvid,
		Storage:       cfg.Storage,
	}
	db := tldb.NewPostgresAdapterFromDBX(f.DBX())
	fr, fe := importer.MainImportFeedVersion(db, opts)
	if fe != nil {
		return nil, fe
	}
	mr := model.FeedVersionImportResult{
		Success: fr.FeedVersionImport.Success,
	}
	return &mr, nil
}

func (f *Finder) FeedVersionUnimport(ctx context.Context, fvid int) (*model.FeedVersionUnimportResult, error) {
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return nil, err
	}

	db := tldb.NewPostgresAdapterFromDBX(f.DBX())
	if err := db.Tx(func(atx tldb.Adapter) error {
		return unimporter.UnimportFeedVersion(atx, fvid, nil)
	}); err != nil {
		return nil, err
	}
	mr := model.FeedVersionUnimportResult{
		Success: true,
	}
	return &mr, nil
}

func (f *Finder) FeedVersionUpdate(ctx context.Context, values model.FeedVersionSetInput) (int, error) {
	if values.ID == nil {
		return 0, errors.New("id required")
	}
	fvid := *values.ID
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return 0, err
	}

	db := tldb.NewPostgresAdapterFromDBX(f.DBX())
	err := db.Tx(func(atx tldb.Adapter) error {
		fv := tl.FeedVersion{}
		fv.ID = fvid
		var cols []string
		if values.Name != nil {
			fv.Name = tt.NewString(*values.Name)
			cols = append(cols, "name")
		} else {
			fv.Name.Valid = false
		}
		if values.Description != nil {
			fv.Description = tt.NewString(*values.Description)
			cols = append(cols, "description")
		} else {
			fv.Description.Valid = false
		}
		return atx.Update(&fv, cols...)
	})
	if err != nil {
		return 0, err
	}
	return fvid, nil
}

func (f *Finder) FeedVersionDelete(ctx context.Context, fvid int) (*model.FeedVersionDeleteResult, error) {
	if err := checkFeedEdit(ctx, fvid); err != nil {
		return nil, err
	}
	return nil, errors.New("temporarily unavailable")
}
