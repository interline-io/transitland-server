package actions

import (
	"context"
	"errors"

	"github.com/interline-io/transitland-lib/dmfr/importer"
	"github.com/interline-io/transitland-lib/dmfr/unimporter"
	"github.com/interline-io/transitland-lib/tl"
	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-lib/tldb"
	"github.com/interline-io/transitland-server/auth"
	"github.com/interline-io/transitland-server/authz"
	"github.com/interline-io/transitland-server/config"
	"github.com/interline-io/transitland-server/model"
)

func FeedVersionImport(ctx context.Context, cfg config.Config, dbf model.Finder, checker *authz.Checker, user auth.User, fvid int) (*model.FeedVersionImportResult, error) {
	if checker != nil {
		if check, err := checker.FeedVersionPermissions(ctx, &authz.FeedVersionRequest{Id: int64(fvid)}); err != nil {
			return nil, err
		} else if !check.Actions.CanEdit {
			return nil, authz.ErrUnauthorized
		}
	}
	opts := importer.Options{
		FeedVersionID: fvid,
		Storage:       cfg.Storage,
	}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
	fr, fe := importer.MainImportFeedVersion(db, opts)
	if fe != nil {
		return nil, fe
	}
	mr := model.FeedVersionImportResult{
		Success: fr.FeedVersionImport.Success,
	}
	return &mr, nil
}

func FeedVersionUnimport(ctx context.Context, cfg config.Config, dbf model.Finder, checker *authz.Checker, user auth.User, fvid int) (*model.FeedVersionUnimportResult, error) {
	if checker != nil {
		if check, err := checker.FeedVersionPermissions(ctx, &authz.FeedVersionRequest{Id: int64(fvid)}); err != nil {
			return nil, err
		} else if !check.Actions.CanEdit {
			return nil, authz.ErrUnauthorized
		}
	}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
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

func FeedVersionUpdate(ctx context.Context, cfg config.Config, dbf model.Finder, checker *authz.Checker, user auth.User, fvid int, values model.FeedVersionSetInput) error {
	if checker != nil {
		if check, err := checker.FeedVersionPermissions(ctx, &authz.FeedVersionRequest{Id: int64(fvid)}); err != nil {
			return err
		} else if !check.Actions.CanEdit {
			return authz.ErrUnauthorized
		}
	}
	db := tldb.NewPostgresAdapterFromDBX(dbf.DBX())
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
		return err
	}
	return nil
}

func FeedVersionDelete(ctx context.Context, cfg config.Config, dbf model.Finder, checker *authz.Checker, user auth.User, fvid int) (*model.FeedVersionDeleteResult, error) {
	if checker != nil {
		if check, err := checker.FeedVersionPermissions(ctx, &authz.FeedVersionRequest{Id: int64(fvid)}); err != nil {
			return nil, err
		} else if !check.Actions.CanEdit {
			return nil, authz.ErrUnauthorized
		}
	}
	return nil, errors.New("temporarily unavailable")
}
