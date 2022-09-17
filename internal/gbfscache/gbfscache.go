package gbfscache

import (
	"context"
	"time"

	"github.com/interline-io/transitland-server/internal/ecache"
	"github.com/interline-io/transitland-server/internal/gbfs"
)

type GbfsFinder struct {
	cache *ecache.Cache[gbfs.GbfsFeed]
	ttl   time.Duration
}

func NewGbfsFinder() *GbfsFinder {
	c := ecache.NewCache[gbfs.GbfsFeed](nil, "gbfs")
	return &GbfsFinder{cache: c}
}

func (c *GbfsFinder) AddData(ctx context.Context, topic string, sf gbfs.GbfsFeed) error {
	c.cache.SetTTL(ctx, topic, sf, c.ttl, c.ttl)
	return nil
}

func (c *GbfsFinder) FindBikes() {

}
