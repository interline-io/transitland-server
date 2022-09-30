package gbfsfinder

import (
	"context"
	"fmt"
	"sort"
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
	ttlRecheck       time.Duration
	ttlExpire        time.Duration
	prefix           string
	bikeSearchKey    string
	stationSearchKey string
}

func NewFinder(client *redis.Client) *Finder {
	c := ecache.NewCache[gbfs.GbfsFeed](client, "gbfs")
	return &Finder{
		ttlRecheck:       5 * time.Minute,
		ttlExpire:        24 * time.Hour,
		cache:            c,
		client:           client,
		prefix:           "gbfs",
		bikeSearchKey:    fmt.Sprintf("%s:bikes", "gbfs"),
		stationSearchKey: fmt.Sprintf("%s:stations", "gbfs"),
	}
}

func (c *Finder) AddData(ctx context.Context, topic string, sf gbfs.GbfsFeed) error {
	// Save basic data
	if err := c.cache.SetTTL(ctx, topic, sf, c.ttlRecheck, c.ttlExpire); err != nil {
		return err
	}
	// Geosearch index bikes
	ts := time.Now().Unix()
	_ = ts
	if c.client != nil {
		var locs []*redis.GeoLocation
		for _, ent := range sf.Bikes {
			locs = append(locs, &redis.GeoLocation{
				Name:      fmt.Sprintf("%s:%s", topic, ent.BikeID.Val),
				Longitude: ent.Lon.Val,
				Latitude:  ent.Lat.Val,
			})
		}
		if err := c.client.GeoAdd(ctx, c.bikeSearchKey, locs...).Err(); err != nil {
			return err
		}
	}
	// Geosearch index docks
	if c.client != nil {
		var locs []*redis.GeoLocation
		for _, ent := range sf.StationInformation {
			locs = append(locs, &redis.GeoLocation{
				Name:      fmt.Sprintf("%s:%s", topic, ent.StationID.Val),
				Longitude: ent.Lon.Val,
				Latitude:  ent.Lat.Val,
			})
		}
		if err := c.client.GeoAdd(ctx, c.stationSearchKey, locs...).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Finder) FindBikes(ctx context.Context, limit *int, where *model.GbfsBikeRequest) ([]*model.GbfsFreeBikeStatus, error) {
	if where == nil || where.Near == nil {
		return nil, nil
	}
	where.Near.Radius = checkFloat(&where.Near.Radius, 0, 10_000)
	pt := *where.Near
	topicKeys, err := c.geosearch(ctx, c.bikeSearchKey, pt)
	if err != nil {
		return nil, err
	}
	var ret []*model.GbfsFreeBikeStatus
	for _, topicKey := range topicKeys {
		sf, ok := c.cache.Get(ctx, topicKey)
		if !ok {
			continue
		}
		for _, ent := range sf.Bikes {
			if d := xy.DistanceHaversine(pt.Lon, pt.Lat, ent.Lon.Val, ent.Lat.Val); d > pt.Radius {
				continue
			}
			b := model.GbfsFreeBikeStatus{
				FreeBikeStatus: ent,
				Feed:           &model.GbfsFeed{GbfsFeed: &sf},
			}
			ret = append(ret, &b)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].BikeID.Val < ret[j].BikeID.Val
	})
	if limit != nil && len(ret) > *limit {
		ret = ret[0:*limit]
	}
	return ret, nil
}

func (c *Finder) FindDocks(ctx context.Context, limit *int, where *model.GbfsDockRequest) ([]*model.GbfsStationInformation, error) {
	if where == nil || where.Near == nil {
		return nil, nil
	}
	where.Near.Radius = checkFloat(&where.Near.Radius, 0, 10_000)
	pt := *where.Near
	topicKeys, err := c.geosearch(ctx, c.stationSearchKey, pt)
	if err != nil {
		return nil, err
	}
	var ret []*model.GbfsStationInformation
	for _, topicKey := range topicKeys {
		sf, ok := c.cache.Get(ctx, topicKey)
		if !ok {
			continue
		}
		for _, ent := range sf.StationInformation {
			if d := xy.DistanceHaversine(pt.Lon, pt.Lat, ent.Lon.Val, ent.Lat.Val); d > pt.Radius {
				continue
			}
			b := model.GbfsStationInformation{
				StationInformation: ent,
				Feed:               &model.GbfsFeed{GbfsFeed: &sf},
			}
			ret = append(ret, &b)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].StationID.Val < ret[j].StationID.Val
	})
	if limit != nil && len(ret) > *limit {
		ret = ret[0:*limit]
	}
	return ret, nil
}

func (c *Finder) geosearch(ctx context.Context, key string, pt model.PointRadius) ([]string, error) {
	topicKeys := map[string][]string{}
	if c.client != nil {
		q := redis.GeoRadiusQuery{
			Radius: pt.Radius,
			Unit:   "m",
		}
		cmd := c.client.GeoRadius(
			ctx,
			key,
			pt.Lon,
			pt.Lat,
			&q,
		)
		locs, err := cmd.Result()
		if err != nil {
			return nil, err
		}
		for _, loc := range locs {
			topic := strings.Split(loc.Name, ":")
			if len(topic) < 4 {
				continue
			}
			topicKey := fmt.Sprintf("%s:%s", topic[0], topic[1])
			elemId := topic[2]
			topicKeys[topicKey] = append(topicKeys[topicKey], elemId)
		}
	} else {
		// If not using redis, get local keys. This is not perfect.
		for _, k := range c.cache.LocalKeys() {
			topicKeys[k] = append(topicKeys[k], "")
		}
	}
	var ret []string
	for k := range topicKeys {
		ret = append(ret, k)
	}
	return ret, nil
}

func checkFloat(v *float64, min float64, max float64) float64 {
	if v == nil || *v < min {
		return min
	} else if *v > max {
		return max
	}
	return *v
}
