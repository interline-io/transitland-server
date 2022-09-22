package gbfsfinder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/interline-io/transitland-server/internal/gbfs"
	"github.com/interline-io/transitland-server/internal/xy"
	"github.com/interline-io/transitland-server/model"
)

type Finder struct {
	client           *redis.Client
	cache            *ecache.Cache[gbfs.GbfsFeed]
	ttl              time.Duration
	prefix           string
	bikeSearchKey    string
	stationSearchKey string
}

func NewFinder(client *redis.Client) *Finder {
	c := ecache.NewCache[gbfs.GbfsFeed](client, "gbfs")
	return &Finder{
		ttl:              1 * time.Hour,
		cache:            c,
		client:           client,
		prefix:           "gbfs",
		bikeSearchKey:    fmt.Sprintf("%s:bikes", "gbfs"),
		stationSearchKey: fmt.Sprintf("%s:stations", "gbfs"),
	}
}

func (c *Finder) AddData(ctx context.Context, topic string, sf gbfs.GbfsFeed) error {
	// Save basic data
	if err := c.cache.SetTTL(ctx, topic, sf, c.ttl, c.ttl); err != nil {
		return err
	}
	// Geosearch index bikes and stations
	if c.client != nil {
		ts := time.Now().Unix()
		var locs []*redis.GeoLocation
		for _, bike := range sf.Bikes {
			locs = append(locs, &redis.GeoLocation{
				Name:      fmt.Sprintf("%s:%s:%d", topic, bike.BikeID.Val, ts),
				Longitude: bike.Lon.Val,
				Latitude:  bike.Lat.Val,
			})
		}
		if err := c.client.GeoAdd(ctx, c.bikeSearchKey, locs...).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Finder) FindBikes(ctx context.Context, pt model.PointRadius) ([]*model.GbfsFreeBikeStatus, error) {
	getBikes := map[string][]string{}
	if c.client != nil {
		q := redis.GeoRadiusQuery{
			Radius: pt.Radius,
			Unit:   "m",
		}
		cmd := c.client.GeoRadius(
			ctx,
			c.bikeSearchKey,
			pt.Lon, pt.Lat, &q,
		)
		locs, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		for _, loc := range locs {
			bikeKey := strings.Split(loc.Name, ":")
			if len(bikeKey) < 4 {
				continue
			}
			bikeTopic := fmt.Sprintf("%s:%s", bikeKey[0], bikeKey[1])
			bikeId := bikeKey[2]
			getBikes[bikeTopic] = append(getBikes[bikeTopic], bikeId)
		}
	} else {
		// If not using redis, get local keys. This is not perfect.
		for _, k := range c.cache.LocalKeys() {
			getBikes[k] = append(getBikes[k], "")
		}
	}
	var ret []*model.GbfsFreeBikeStatus
	for bikeTopic := range getBikes {
		// Check bikes in matched topics, redo distance check
		sf, ok := c.cache.Get(ctx, bikeTopic)
		if !ok {
			continue
		}
		for _, bike := range sf.Bikes {
			bikeDist := xy.DistanceHaversine(pt.Lon, pt.Lat, bike.Lon.Val, bike.Lat.Val)
			if bikeDist > pt.Radius {
				continue
			}
			b := model.GbfsFreeBikeStatus{
				FreeBikeStatus: bike,
			}
			ret = append(ret, &b)
		}
	}
	return ret, nil
}

func (c *Finder) FindStations(pt model.PointRadius) {

}
