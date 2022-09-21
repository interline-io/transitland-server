package gbfsfinder

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/interline-io/transitland-server/internal/gbfs"
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
		cache:            c,
		client:           client,
		prefix:           "gbfs",
		bikeSearchKey:    fmt.Sprintf("%s:bikes", "gbfs"),
		stationSearchKey: fmt.Sprintf("%s:stations", "gbfs"),
	}
}

func (c *Finder) AddData(ctx context.Context, topic string, sf gbfs.GbfsFeed) error {
	// Save basic data
	c.cache.SetTTL(ctx, topic, sf, 1*time.Hour, 1*time.Hour)
	// Geo index bikes and stations
	ts := time.Now().UnixNano()
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
	c.cache.SetTTL(ctx, topic, sf, c.ttl, c.ttl)
	return nil
}

func (c *Finder) FindBikes(ctx context.Context, pt model.PointRadius) ([]*model.GbfsFreeBikeStatus, error) {
	q := redis.GeoRadiusQuery{
		Radius:    pt.Radius,
		Unit:      "m",
		WithDist:  true,
		WithCoord: true,
	}
	cmd := c.client.GeoRadius(
		ctx,
		c.bikeSearchKey,
		pt.Lon, pt.Lat, &q,
	)
	locs, err := cmd.Result()
	if err != nil {
		panic(err)
	}
	getBikes := map[string][]string{}
	for _, loc := range locs {
		bikeKey := strings.Split(loc.Name, ":")
		if len(bikeKey) < 4 {
			continue
		}
		bikeTopic := fmt.Sprintf("%s:%s", bikeKey[0], bikeKey[1])
		bikeId := bikeKey[2]
		getBikes[bikeTopic] = append(getBikes[bikeTopic], bikeId)
	}
	var ret []*model.GbfsFreeBikeStatus
	for bikeTopic, bikeIds := range getBikes {
		bikeSet := map[string]bool{}
		for _, b := range bikeIds {
			bikeSet[b] = true
		}
		sf, ok := c.cache.Get(ctx, bikeTopic)
		if !ok {
			continue
		}
		for _, bike := range sf.Bikes {
			if bikeSet[bike.BikeID.Val] {
				b := model.GbfsFreeBikeStatus{
					FreeBikeStatus: bike,
				}
				ret = append(ret, &b)
			}
		}
	}
	return ret, nil
}

func (c *Finder) FindStations(pt model.PointRadius) {

}

func uniq(v []string) []string {
	m := map[string]bool{}
	for _, k := range v {
		m[k] = true
	}
	var ret []string
	for k := range m {
		ret = append(ret, k)
	}
	return ret
}
