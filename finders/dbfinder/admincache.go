package dbfinder

import (
	"context"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/xy"

	"github.com/interline-io/transitland-dbutil/dbutil"
	"github.com/interline-io/transitland-lib/tlxy"
	"github.com/interline-io/transitland-lib/tt"
	"github.com/jmoiron/sqlx"
	"github.com/tidwall/rtree"
)

type adminCacheItem struct {
	Adm0Name string
	Adm1Name string
	Adm0Iso  string
	Adm1Iso  string
	Geometry *geom.Polygon
}

type adminCache struct {
	lock  sync.Mutex
	index rtree.Generic[*adminCacheItem]
	cache map[tlxy.Point]*adminCacheItem
}

func newAdminCache() *adminCache {
	return &adminCache{
		cache: map[tlxy.Point]*adminCacheItem{},
	}
}

func (c *adminCache) LoadAdmins(ctx context.Context, dbx sqlx.Ext) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	var ents []struct {
		Adm0Name tt.String
		Adm1Name tt.String
		Adm0Iso  tt.String
		Adm1Iso  tt.String
		Geometry tt.Geometry
	}
	q := sq.Select(
		"ne.name as adm1_name",
		"ne.admin as adm0_name",
		"iso_a2 as adm0_iso",
		"iso_3166_2 as adm1_iso",
		"ne.geometry",
	).
		From("ne_10m_admin_1_states_provinces ne")
	if err := dbutil.Select(ctx, dbx, q, &ents); err != nil {
		return err
	}
	for _, ent := range ents {
		g, ok := ent.Geometry.Val.(*geom.MultiPolygon)
		if !ok {
			continue
		}
		for i := 0; i < g.NumPolygons(); i++ {
			item := adminCacheItem{
				Adm0Name: ent.Adm0Name.Val,
				Adm1Name: ent.Adm1Name.Val,
				Adm0Iso:  ent.Adm0Iso.Val,
				Adm1Iso:  ent.Adm1Iso.Val,
				Geometry: g.Polygon(i),
			}
			bbox := item.Geometry.Bounds()
			b1 := [2]float64{bbox.Min(0), bbox.Min(1)}
			b2 := [2]float64{bbox.Max(0), bbox.Max(1)}
			c.index.Insert(b1, b2, &item)
		}
	}
	return nil
}

func (c *adminCache) Check(pt tlxy.Point) (adminCacheItem, bool) {
	ret, count := c.CheckPolygon(pt)
	if count >= 1 {
		return ret, count == 1
	}
	tolerance := 0.25
	nearestAdmin, _, count := c.NearestPolygon(pt, tolerance)
	// fmt.Println("nearestPolygon:", pt.Lon, pt.Lat, "admin:", nearestAdmin, "count:", count)
	return nearestAdmin, count >= 1
}

func (c *adminCache) CheckPolygon(p tlxy.Point) (adminCacheItem, int) {
	// Checking just the index can be much faster, but can be invalid in open water, e.g. 0,0 = Kiribati
	// However, in practice, most land area on Earth falls into more than 1 admin bbox
	// No, we are not being fancy with projections.
	// That could be improved.
	ret := adminCacheItem{}
	gp := geom.NewPointFlat(geom.XY, []float64{p.Lon, p.Lat})
	count := 0
	c.index.Search(
		[2]float64{p.Lon, p.Lat},
		[2]float64{p.Lon, p.Lat},
		func(min, max [2]float64, s *adminCacheItem) bool {
			if pointInPolygon(s.Geometry, gp) {
				ret.Adm0Name = s.Adm0Name
				ret.Adm1Name = s.Adm1Name
				ret.Adm0Iso = s.Adm0Iso
				ret.Adm1Iso = s.Adm1Iso
				count += 1
			}
			return true
		},
	)
	return ret, count
}

func (c *adminCache) NearestPolygon(p tlxy.Point, tolerance float64) (adminCacheItem, float64, int) {
	ret := adminCacheItem{}
	minDist := -1.0
	gp := geom.NewPointFlat(geom.XY, []float64{p.Lon, p.Lat})
	count := 0
	c.index.Search(
		[2]float64{p.Lon - tolerance, p.Lat - tolerance},
		[2]float64{p.Lon + tolerance, p.Lat + tolerance},
		func(min, max [2]float64, s *adminCacheItem) bool {
			d := pointPolygonDistance(s.Geometry, gp)
			if d < tolerance && (d < minDist || minDist < 0) {
				ret.Adm0Name = s.Adm0Name
				ret.Adm1Name = s.Adm1Name
				ret.Adm0Iso = s.Adm0Iso
				ret.Adm1Iso = s.Adm1Iso
				count += 1
				minDist = d
			}
			return true
		},
	)
	return ret, minDist, count
}

func pointInPolygon(pg *geom.Polygon, p *geom.Point) bool {
	if !xy.IsPointInRing(geom.XY, p.Coords(), pg.LinearRing(0).FlatCoords()) {
		return false
	}
	for i := 1; i < pg.NumLinearRings(); i++ {
		if xy.IsPointInRing(geom.XY, p.Coords(), pg.LinearRing(i).FlatCoords()) {
			return false
		}
	}
	return true
}

func pointPolygonDistance(pg *geom.Polygon, p *geom.Point) float64 {
	minDist := -1.0
	c := geom.Coord{p.X(), p.Y()}
	for i := 0; i < pg.NumLinearRings(); i++ {
		d := xy.DistanceFromPointToLineString(p.Layout(), c, pg.LinearRing(i).FlatCoords())
		if d < minDist || minDist < 0 {
			minDist = d
		}
	}
	return minDist
}
