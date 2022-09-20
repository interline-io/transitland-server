package gbfscache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/model"
)

type GbfsFinder struct {
	client *redis.Client
	cache  *ecache.Cache[gbfs.GbfsFeed]
	ttl    time.Duration
}

func NewGbfsFinder(client *redis.Client) *GbfsFinder {
	c := ecache.NewCache[gbfs.GbfsFeed](nil, "gbfs")
	return &GbfsFinder{cache: c, client: client}
}

func (c *GbfsFinder) AddData(ctx context.Context, topic string, sf gbfs.GbfsFeed) error {
	ts := 0
	var locs []*redis.GeoLocation
	for _, bike := range sf.Bikes {
		locs = append(locs, &redis.GeoLocation{
			Name:      fmt.Sprintf("%s:%s:%d", topic, bike.BikeID.Val, ts),
			Longitude: bike.Lon.Val,
			Latitude:  bike.Lat.Val,
		})
	}
	if err := c.client.GeoAdd(ctx, "gbfs:bikes", locs...).Err(); err != nil {
		return err
	}
	c.cache.SetTTL(ctx, topic, sf, c.ttl, c.ttl)
	return nil
}

func (c *GbfsFinder) FindBikes(ctx context.Context, pt model.PointRadius) {
	q := redis.GeoRadiusQuery{
		Radius:    pt.Radius,
		Unit:      "m",
		WithDist:  true,
		WithCoord: true,
	}
	cmd := c.client.GeoRadius(
		ctx,
		"gbfs:bikes",
		pt.Lon, pt.Lat, &q,
	)
	locs, err := cmd.Result()
	if err != nil {
		panic(err)
	}
	for _, loc := range locs {
		fmt.Println("loc:", loc, loc.Dist, loc.Longitude, loc.Latitude)
	}
}

func (c *GbfsFinder) FindStations(pt model.PointRadius) {

}
