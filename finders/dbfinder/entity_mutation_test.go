package dbfinder

// TODO: testconfig is an import cycle.
// func TestCreateStop(t *testing.T) {
// 	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
// 		finder := cfg.Finder
// 		ctx := model.WithConfig(context.Background(), cfg)
// 		fv := model.FeedVersionInput{ID: toPtr(1)}
// 		stopInput := model.StopInput{
// 			FeedVersion: &fv,
// 			StopID:      toPtr(fmt.Sprintf("%d", time.Now().UnixNano())),
// 			StopName:    toPtr("hello"),
// 			Geometry:    toPtr(tt.NewPoint(-122.271604, 37.803664)),
// 		}
// 		eid, err := finder.CreateStop(ctx, stopInput)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		checkEnt := tl.Stop{}
// 		checkEnt.ID = eid
// 		atx := tldb.NewPostgresAdapterFromDBX(cfg.Finder.DBX())
// 		if err := atx.Find(&checkEnt); err != nil {
// 			t.Fatal(err)
// 		}
// 		assert.Equal(t, stopInput.StopID, &checkEnt.StopID)
// 		assert.Equal(t, stopInput.StopName, &checkEnt.StopName)
// 		assert.Equal(t, stopInput.Geometry.Coords(), checkEnt.Geometry.Coords())
// 	})
// }

// func TestUpdateStop(t *testing.T) {
// 	testconfig.ConfigTxRollback(t, testconfig.Options{}, func(cfg model.Config) {
// 		finder := cfg.Finder
// 		ctx := model.WithConfig(context.Background(), cfg)
// 		fv := model.FeedVersionInput{ID: toPtr(1)}
// 		stopInput := model.StopInput{
// 			FeedVersion: &fv,
// 			StopID:      toPtr(fmt.Sprintf("%d", time.Now().UnixNano())),
// 			StopName:    toPtr("hello"),
// 			Geometry:    toPtr(tt.NewPoint(-122.271604, 37.803664)),
// 		}
// 		eid, err := finder.CreateStop(ctx, stopInput)
// 		if err != nil {
// 			t.Fatal(err)
// 		}
// 		stopUpdate := model.StopInput{
// 			ID:       toPtr(eid),
// 			StopID:   toPtr(fmt.Sprintf("update-%d", time.Now().UnixNano())),
// 			Geometry: toPtr(tt.NewPoint(-122.0, 38.0)),
// 		}
// 		if _, err := finder.UpdateStop(ctx, stopUpdate); err != nil {
// 			t.Fatal(err)
// 		}
// 		checkEnt := tl.Stop{}
// 		checkEnt.ID = eid
// 		atx := tldb.NewPostgresAdapterFromDBX(cfg.Finder.DBX())
// 		if err := atx.Find(&checkEnt); err != nil {
// 			t.Fatal(err)
// 		}
// 		assert.Equal(t, stopUpdate.StopID, &checkEnt.StopID)
// 		assert.Equal(t, stopUpdate.Geometry.Coords(), checkEnt.Geometry.Coords())
// 	})
// }
