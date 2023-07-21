package admincache

import (
	"context"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/twpayne/go-geom"
	geomxy "github.com/twpayne/go-geom/xy"

	"github.com/interline-io/transitland-lib/tl/tt"
	"github.com/interline-io/transitland-server/internal/dbutil"
	"github.com/interline-io/transitland-server/internal/xy"
	"github.com/jmoiron/sqlx"
	"github.com/tidwall/rtree"
)

type AdminItem struct {
	Adm0Name string
	Adm1Name string
	Adm0Iso  string
	Adm1Iso  string
	Geometry *geom.Polygon
	Count    int
}

type AdminCache struct {
	lock  sync.Mutex
	index rtree.Generic[*AdminItem]
	cache map[xy.Point]*AdminItem
}

func NewAdminCache() *AdminCache {
	return &AdminCache{
		cache: map[xy.Point]*AdminItem{},
	}
}

func (c *AdminCache) LoadAdmins(ctx context.Context, dbx sqlx.Ext) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// Don't have a tt.MultiPolygon to handle encoding/decoding
	var ents []struct {
		Adm0Name tt.String
		Adm1Name tt.String
		Adm0Iso  tt.String
		Adm1Iso  tt.String
		Geometry tt.Geometry
	}
	q := sq.Select("ne.name as adm1_name", "ne.admin as adm0_name", "iso_a2 as adm0_iso", "iso_3166_2 as adm1_iso", "ne.geometry").From("ne_10m_admin_1_states_provinces ne")
	if err := dbutil.Select(ctx, dbx, q, &ents); err != nil {
		return err
	}
	for _, ent := range ents {
		g, ok := ent.Geometry.Geometry.(*geom.MultiPolygon)
		if !ok {
			continue
		}
		for i := 0; i < g.NumPolygons(); i++ {
			item := AdminItem{
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

func (c *AdminCache) CheckPoints(pts []xy.Point) ([]AdminItem, error) {
	ret := make([]AdminItem, len(pts))
	for i, pt := range pts {
		ret[i] = c.Check(pt)
	}
	return ret, nil
}

func (c *AdminCache) Check(pt xy.Point) AdminItem {
	// This can be much faster, but can be invalid in open water, e.g. 0,0 = Kiribati
	// However, in practice, most land area on Earth falls into more than 1 admin bbox
	// if a, _ := c.CheckIndex(pt); a.Count < 2 {
	// 	fmt.Println("index ok")
	// 	return a
	// }
	a, _ := c.CheckPolygon(pt)
	return a
}

func (c *AdminCache) CheckIndex(p xy.Point) (AdminItem, int) {
	ret := AdminItem{}
	count := 0
	c.index.Search(
		[2]float64{p.Lon, p.Lat},
		[2]float64{p.Lon, p.Lat},
		func(min, max [2]float64, s *AdminItem) bool {
			ret.Adm0Name = s.Adm0Name
			ret.Adm1Name = s.Adm1Name
			ret.Adm0Iso = s.Adm0Iso
			ret.Adm1Iso = s.Adm1Iso
			ret.Count += 1
			count += 1
			return true
		},
	)
	return ret, count
}

func (c *AdminCache) CheckPolygon(p xy.Point) (AdminItem, int) {
	// No, we are not being fancy with projections.
	// That could be improved.
	ret := AdminItem{}
	gp := geom.NewPointFlat(geom.XY, []float64{p.Lon, p.Lat})
	count := 0
	c.index.Search(
		[2]float64{p.Lon, p.Lat},
		[2]float64{p.Lon, p.Lat},
		func(min, max [2]float64, s *AdminItem) bool {
			if pointInPolygon(s.Geometry, gp) {
				ret.Adm0Name = s.Adm0Name
				ret.Adm1Name = s.Adm1Name
				ret.Adm0Iso = s.Adm0Iso
				ret.Adm1Iso = s.Adm1Iso
				ret.Count += 1
				count += 1
			}
			return true
		},
	)
	return ret, count
}

func pointInPolygon(pg *geom.Polygon, p *geom.Point) bool {
	if !geomxy.IsPointInRing(geom.XY, p.Coords(), pg.LinearRing(0).FlatCoords()) {
		return false
	}
	for i := 1; i < pg.NumLinearRings(); i++ {
		if geomxy.IsPointInRing(geom.XY, p.Coords(), pg.LinearRing(i).FlatCoords()) {
			return false
		}
	}
	return true
}
